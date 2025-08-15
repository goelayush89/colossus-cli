🚀 COLOSSUS CLI v1.1 - Windows Binary
======================================

This folder contains the Windows executable for Colossus CLI, a powerful
Go-based alternative to Ollama for running large language models locally.

✨ NEW in v1.1:
- 📊 Visual progress bars for model downloads
- ⚡ Real-time download speed and ETA display
- 🚀 Enhanced GGUF model repositories with 7 guaranteed models
- 🤖 REAL AI INFERENCE: Defaults to llama.cpp (no more simulation!)
- 🔥 Production-ready with actual model responses

📁 FILES:
----------
- colossus.exe        : Main executable
- run-colossus.bat    : Interactive menu launcher
- start-server.bat    : Quick server start
- README.txt          : This file

🎯 HOW TO USE:
--------------

Option 1: Double-click "run-colossus.bat" for an interactive menu

Option 2: Double-click "start-server.bat" to quickly start the API server

Option 3: Use Command Prompt/PowerShell:
   cd "path\to\this\folder"
   colossus.exe --help

🚀 QUICK START:
---------------
1. Double-click "run-colossus.bat"
2. Choose option 1 to start the server
3. Choose option 5 to download a model (try "qwen" - smallest)
4. Choose option 6 to start chatting

📦 AVAILABLE GGUF MODELS:
-------------------------
- qwen      (0.5B) - Smallest, fastest, great for testing
- tinyllama (1.1B) - Popular lightweight model
- gemma     (2B)   - Google's Gemma model
- phi       (2.7B) - Microsoft's Phi model  
- mistral   (7B)   - Larger, more capable
- llama2    (7B)   - Meta's Llama-2
- codellama (7B)   - Specialized for coding tasks

🌐 API USAGE:
-------------
Once the server is running, you can use it via HTTP:

# List models
curl http://localhost:11434/api/tags

# Generate text
curl -X POST http://localhost:11434/api/generate ^
  -H "Content-Type: application/json" ^
  -d "{\"model\": \"tinyllama\", \"prompt\": \"Hello world\"}"

# Chat
curl -X POST http://localhost:11434/api/chat ^
  -H "Content-Type: application/json" ^
  -d "{\"model\": \"tinyllama\", \"messages\": [{\"role\": \"user\", \"content\": \"Hello\"}]}"

✨ FEATURES:
------------
- 🔥 Full Ollama API compatibility
- ⚡ Real inference with llama.cpp integration  
- 🚀 GPU acceleration (CUDA/ROCm/Metal)
- 🤗 Hugging Face Hub integration
- 📦 Automatic model management
- 🛡️ Model format validation
- 💬 Interactive chat sessions
- 🌐 REST API for integrations

🆘 TROUBLESHOOTING:
-------------------
- If you see "This is a command line tool", use the .bat files or command prompt
- If server fails to start, check if port 11434 is available
- For GPU acceleration, ensure proper drivers are installed
- Models are downloaded to: %USERPROFILE%\.colossus\models

📖 MORE INFO:
-------------
GitHub: https://github.com/yourusername/colossus-cli
Documentation: https://yourusername.github.io/colossus-cli

Built with ❤️ in Go
