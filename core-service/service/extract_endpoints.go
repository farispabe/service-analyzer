package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ExtractEndpointsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request JSON
	var request struct {
		RepositoryURL string `json:"repository_url"`
		Pattern       string `json:"pattern"`
		AccessToken   string `json:"access_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.RepositoryURL == "" {
		http.Error(w, "Repository URL is required", http.StatusBadRequest)
		return
	}

	if request.Pattern == "" {
		http.Error(w, "Pattern is required", http.StatusBadRequest)
		return
	}

	// Add token to the repository URL
	authenticatedRepoURL := addAccessTokenToURL(request.RepositoryURL, request.AccessToken)
	fmt.Println("authenticated repo url : ", authenticatedRepoURL)

	// Clone the repository
	repoRoot, err := cloneRepository(authenticatedRepoURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to clone repository: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(repoRoot) // Clean up cloned repository after use

	// Extract endpoints
	endpoints, err := extractEndpoints(repoRoot)
	if err != nil {
		fmt.Printf("Error extracting endpoints: %v\n", err)
		return
	}

	// Respond with extracted endpoints
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(endpoints)
}

func addAccessTokenToURL(repoURL, accessToken string) string {
	// Parse the URL to split protocol and the rest of the URL
	parts := strings.Split(repoURL, "://")
	if len(parts) != 2 {
		return repoURL // Return original URL if it doesn't match expected format
	}
	// Insert the token before the domain
	return fmt.Sprintf("%s://oauth2:%s@%s", parts[0], accessToken, parts[1])
}

func cloneRepository(repoURL string) (string, error) {
	// Derive the project name from the repository URL
	parts := strings.Split(repoURL, "/")
	projectName := strings.TrimSuffix(parts[len(parts)-1], ".git")

	// Clone the repository into /tmp/repo/<projectName>
	repoRoot := filepath.Join("/tmp/repo", projectName)
	err := os.MkdirAll("/tmp/repo", os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Use git clone with authentication
	cmd := exec.Command("git", "clone", repoURL, repoRoot)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	return repoRoot, nil
}

func extractEndpoints(repoRoot string) ([]Endpoint, error) {
	var endpoints []Endpoint

	// Pattern 1: ./internal/handler
	handlerPath := filepath.Join(repoRoot, "internal", "handler")
	handlerEndpoints, err := processHandlerInterfaceFiles(handlerPath, filepath.Join(repoRoot, "internal", "usecase"))
	if err != nil {
		return nil, fmt.Errorf("error processing internal handler: %w", err)
	}
	endpoints = append(endpoints, handlerEndpoints...)

	// Pattern 2: ./service
	servicePath := filepath.Join(repoRoot, "service")
	serviceEndpoints, err := processInterfaceFiles(servicePath, servicePath)
	if err != nil {
		return nil, fmt.Errorf("error processing service: %w", err)
	}
	endpoints = append(endpoints, serviceEndpoints...)

	return endpoints, nil
}

func processHandlerInterfaceFiles(handlerPath, implementationPath string) ([]Endpoint, error) {
	var endpoints []Endpoint

	if _, err := os.Stat(handlerPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("handler path does not exist: %s", handlerPath)
	}

	err := filepath.Walk(handlerPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), "init.go") {
			fileEndpoints, err := parseInterfaceFile(path, implementationPath)
			if err != nil {
				return fmt.Errorf("error parsing interface file %s: %w", path, err)
			}
			endpoints = append(endpoints, fileEndpoints...)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error during filepath walk: %w", err)
	}

	return endpoints, nil
}

func processInterfaceFiles(interfacePath, implementationPath string) ([]Endpoint, error) {
	var endpoints []Endpoint

	err := filepath.Walk(interfacePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".go") {
			return nil
		}

		fileEndpoints, err := parseInterfaceFile(path, implementationPath)
		if err != nil {
			fmt.Printf("Error parsing interface file %s: %v\n", path, err)
			return nil
		}
		endpoints = append(endpoints, fileEndpoints...)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return endpoints, nil
}

func parseInterfaceFile(interfaceFile, implementationPath string) ([]Endpoint, error) {
	var endpoints []Endpoint

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, interfaceFile, nil, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}

			for _, method := range interfaceType.Methods.List {
				for _, name := range method.Names {
					implementation, localCalls, implFile := findImplementationWithPackageMethods(name.Name, implementationPath)
					endpoints = append(endpoints, Endpoint{
						InterfaceName:      typeSpec.Name.Name,
						MethodName:         name.Name,
						InterfaceFile:      interfaceFile,
						Implementation:     implementation,
						ImplementationFile: implFile,
						LocalCalls:         localCalls,
					})
				}
			}
		}
	}

	return endpoints, nil
}

func findImplementationWithPackageMethods(methodName, implementationPath string) (string, []string, string) {
	var implementation string
	var localCalls []string
	var implFile string

	// Map to store methods in the package
	packageMethods := make(map[string]string)

	// Set to track visited methods to avoid redundant traversal
	visited := make(map[string]bool)

	filepath.Walk(implementationPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			return nil
		}

		// Collect all methods and their code in this package
		ast.Inspect(node, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			code := extractCodeFromAST(funcDecl, fset)
			packageMethods[funcDecl.Name.Name] = code
			return true
		})

		// Find the implementation of the requested method
		ast.Inspect(node, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != methodName {
				return true
			}

			implFile = path
			implementation = extractCodeFromAST(funcDecl, fset)
			localCalls = extractLocalCalls(funcDecl.Body)

			return false
		})

		return nil
	})

	// Add codes of local calls to the implementation (recursively)
	for _, call := range localCalls {
		if _, exists := packageMethods[call]; exists && !visited[call] {
			// Mark this method as visited
			visited[call] = true

			// Recursively add methods called by the current call
			recursiveCalls, _, _ := findImplementationWithPackageMethods(call, implementationPath)
			implementation += "\n\n" + recursiveCalls
		}
	}

	return implementation, localCalls, implFile
}

func extractLocalCalls(body *ast.BlockStmt) []string {
	var calls []string
	if body == nil {
		return calls
	}

	ast.Inspect(body, func(n ast.Node) bool {
		// Check for function or method calls
		if call, ok := n.(*ast.CallExpr); ok {
			switch fun := call.Fun.(type) {
			case *ast.Ident:
				// Direct function call
				calls = append(calls, fun.Name)
			case *ast.SelectorExpr:
				// Method call (e.g., uc.someFunc)
				if _, ok := fun.X.(*ast.Ident); ok {
					calls = append(calls, fun.Sel.Name) // Add only the method name
					// calls = append(calls, ident.Name)
				}
			}
		}
		return true
	})

	return calls
}

func extractCodeFromAST(node ast.Node, fset *token.FileSet) string {
	var buf bytes.Buffer
	err := printer.Fprint(&buf, fset, node)
	if err != nil {
		return ""
	}
	return buf.String()
}
