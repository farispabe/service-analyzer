import os
import json
from flask import Flask, request, jsonify
from flask_cors import CORS
import requests
from dotenv import load_dotenv
import openai
import re

# Load environment variables from .env file
load_dotenv()

# Configure the OpenRouter API key
openai.api_key = os.getenv("OPENROUTER_API_KEY")
openai.api_base = "https://openrouter.ai/api/v1"

app = Flask(__name__)
CORS(app, resources={r"/api/*": {"origins": "http://localhost:7012"}})  # Allow core-service

@app.route('/api/data', methods=['GET'])
def get_data():
    return jsonify({"data": "AI Service Data"})

@app.route('/generate-summary', methods=['POST'])
def generate_summary_and_sequence():
    try:
        # Get the data from the request
        data = request.get_json()

        if not data or 'method_name' not in data or 'method_code' not in data:
            return jsonify({"error": "Method name and method code are required"}), 400

        method_name = data['method_name']
        method_code = data['method_code']
        model = data['model']
        message_type = data['message_type']

        # Construct the prompt for OpenRouter with the method name and its full code
        prompt = f"""
        Please analyze the following Golang method and generate a detailed summary and sequence diagram. The method might call other functions within the same package, and the sequence diagram should capture all interactions and function calls. Provide the output in a structured JSON format with two fields:
        - "summary": A detailed description of the method's purpose, functionality, any important flow or conditions, and a mention of other methods it calls in the same package.
        - "sequence_diagram_code": The sequence diagram code in Mermaid format (without triple backticks), representing the sequence of function calls, conditional branches (like 'if' statements), error handling, and the overall flow. Make sure it is in a valid Mermaid format to be visualized.

        The method should be analyzed in detail, including:
        1. The input parameters to the method and any relevant validation.
        2. Any internal logic and method calls, including calls to other functions in the same package.
        3. The return values, error handling, and any conditions like `if`, `else`, or loops.
        4. The interactions between different components, including any spans or external services involved.

        Method Name: {method_name}
        Method Code:
        {method_code}
        """

        # Determine the message format based on message_type
        if message_type == "plain_text":
            messages = [{"role": "user", "content": prompt}]
        elif message_type == "text_object":
            messages = [{"role": "user", "content": [{"type": "text", "text": prompt}]}]
        else:
            return jsonify({"error": "Invalid engine.message_type. Supported values are 'plain_text' and 'text_object'."}), 400

        # Make a request to OpenRouter
        response = openai.ChatCompletion.create(
            model=model,
            messages=messages
        )

        # Extract and clean the AI's output
        ai_output = response["choices"][0]["message"]["content"].strip()

        # Remove explanations and extra formatting (if any)
        json_match = re.search(r"```json\s*(\{.*?\}|\[.*?\])\s*```", ai_output, re.DOTALL)
        if json_match:
            ai_output = json_match.group(1)  # Extract the JSON content  # Strip ```json ... ```

        # Correct ai output
        print("raw ai output : \n", ai_output, flush=True)
        ai_output_cleaned = ai_output.replace("\n", "\\n").replace("\t", "\\t").replace("\r", "\\r")
        print("escaped ai output : \n", ai_output_cleaned, flush=True)
        ai_output_cleaned = correct_broken_json(ai_output_cleaned)

        try:
            # Parse the AI's output as JSON
            result = json.loads(ai_output_cleaned)
        except json.JSONDecodeError as e:
            return jsonify({"error": f"Error parsing AI response as JSON: {str(e)}"}), 500
        except Exception as e:
            return jsonify({"error": f"Error validating AI response structure: {str(e)}"}), 500

        # Prepare the final response
        return jsonify(result), 200
    
    except Exception as e:
        # Debugging: Log the exception
        print(f"Exception occurred: {str(e)}", flush=True)
        return jsonify({"error": f"Error in categorize_document: {str(e)}"}), 500
    
def correct_broken_json(broken_json_str):
    broken_json_str = re.sub(r'\{\s*\\n|\{\s*\\t|\{\s*\\r', '{', broken_json_str)
    broken_json_str = re.sub(r'\\n\s*\}|\\t\s*\}|\\r\s*\}', '}', broken_json_str)
    broken_json_str = re.sub(r'\"\s*,\s*\\n\s*\"|\"\s*,\s*\\t\s*\"|\"\s*,\s*\\r\s*\"', '\",\n\"', broken_json_str)
    broken_json_str = re.sub(r'\"\s*:\s*\\n\s*\"|\"\s*:\s*\\t\s*\"|\"\s*:\s*\\r\s*\"', '\":\"', broken_json_str)
    print("corrected ai output : \n", broken_json_str, flush=True)
    
    # Step 2: Try to parse the cleaned-up JSON string
    try:
        # Parse the fixed JSON string
        parsed_json = json.loads(broken_json_str)
        # Return a nicely formatted JSON string
        return json.dumps(parsed_json, indent=2)
    except json.JSONDecodeError as e:
        print(f"Error parsing JSON: {e}")
        return None  # Return None if parsing fails

if __name__ == '__main__':
    app.run(host="0.0.0.0", port=7011)
