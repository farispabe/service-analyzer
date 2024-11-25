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
	"path/filepath"
	"strings"
)

func ExtractEndpointsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request JSON
	var request struct {
		RepoRoot string `json:"repo_root"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.RepoRoot == "" {
		http.Error(w, "Repository path is required", http.StatusBadRequest)
		return
	}

	// Extract endpoints
	endpoints, err := extractEndpoints(request.RepoRoot)
	if err != nil {
		fmt.Printf("Error extracting endpoints: %v\n", err)
		return
	}

	fmt.Println("endpoints ", endpoints)
	for _, endpoint := range endpoints {
		fmt.Printf("Interface: %s\nMethod: %s\nInterface File: %s\nImplementation File: %s\nImplementation:\n%s\nLocal Calls: %v\n\n",
			endpoint.InterfaceName, endpoint.MethodName, endpoint.InterfaceFile, endpoint.ImplementationFile, endpoint.Implementation, endpoint.LocalCalls)
	}

	// Respond with extracted endpoints
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(endpoints)
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

	// Check if the handlerPath exists
	if _, err := os.Stat(handlerPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("handler path does not exist: %s", handlerPath)
	}

	// Recursively find init.go files in handlerPath and its subdirectories
	err := filepath.Walk(handlerPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		// Check for "init.go" files
		if !info.IsDir() && strings.HasSuffix(info.Name(), "init.go") {
			fmt.Printf("Processing file: %s\n", path) // Debug log

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

	// Find all Go files in the interface path
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

	// Extract interfaces and methods
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
					implementation, localCalls, implFile := findImplementation(name.Name, implementationPath)
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

func findImplementation(methodName, implementationPath string) (string, []string, string) {
	var implementation string
	var localCalls []string
	var implFile string

	// Walk through files in the implementation path
	filepath.Walk(implementationPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			return nil
		}

		ast.Inspect(node, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != methodName {
				return true
			}

			implFile = path
			implementation = prettyPrintAST(funcDecl.Body, fset)
			localCalls = extractLocalCalls(funcDecl.Body)

			return false
		})

		return nil
	})

	return implementation, localCalls, implFile
}

func extractLocalCalls(body *ast.BlockStmt) []string {
	var calls []string
	if body == nil {
		return calls
	}

	ast.Inspect(body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if ident, ok := call.Fun.(*ast.Ident); ok {
				calls = append(calls, ident.Name)
			}
		}
		return true
	})

	return calls
}

func prettyPrintAST(node ast.Node, fset *token.FileSet) string {
	if node == nil {
		return ""
	}
	var buf bytes.Buffer
	err := printer.Fprint(&buf, fset, node)
	if err != nil {
		return fmt.Sprintf("error printing AST: %v", err)
	}
	return buf.String()
}
