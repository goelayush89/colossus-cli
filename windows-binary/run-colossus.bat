@echo off
REM Colossus CLI - Windows Launcher
REM This batch file makes it easy to run Colossus from File Explorer

echo.
echo ========================================
echo    🚀 Colossus CLI v1.1 - Ollama Alternative
echo ========================================
echo.
echo ✨ NEW: Visual progress bars for downloads!
echo 📊 Real-time download speed and ETA
echo 🚀 Enhanced GGUF model repositories
echo.

REM Check if binary exists
if not exist "%~dp0colossus.exe" (
    echo ❌ Error: colossus.exe not found!
    echo Please ensure colossus.exe is in the same folder as this batch file.
    pause
    exit /b 1
)

REM Show menu
:MENU
echo Choose an option:
echo.
echo 1. Start Server (API mode)
echo 2. List Available Commands
echo 3. Check GPU Information
echo 4. List Models
echo 5. Pull a Model
echo 6. Start Chat Session
echo 7. Exit
echo.
set /p choice="Enter your choice (1-7): "

if "%choice%"=="1" goto START_SERVER
if "%choice%"=="2" goto SHOW_HELP
if "%choice%"=="3" goto GPU_INFO
if "%choice%"=="4" goto LIST_MODELS
if "%choice%"=="5" goto PULL_MODEL
if "%choice%"=="6" goto START_CHAT
if "%choice%"=="7" goto EXIT
echo Invalid choice. Please try again.
goto MENU

:START_SERVER
echo.
echo 🚀 Starting Colossus server...
echo Server will be available at http://localhost:11434
echo Press Ctrl+C to stop the server
echo.
"%~dp0colossus.exe" serve --verbose
goto MENU

:SHOW_HELP
echo.
"%~dp0colossus.exe" --help
echo.
pause
goto MENU

:GPU_INFO
echo.
echo 🔍 Checking GPU information...
"%~dp0colossus.exe" gpu info
echo.
pause
goto MENU

:LIST_MODELS
echo.
echo 📦 Listing available models...
"%~dp0colossus.exe" models list
echo.
pause
goto MENU

:PULL_MODEL
echo.
echo 📦 Available GGUF Models (guaranteed downloads):
echo.
echo   1. qwen      - 0.5B params (smallest, fastest)
echo   2. tinyllama - 1.1B params (popular choice)
echo   3. gemma     - 2B params (Google model)
echo   4. phi       - 2.7B params (Microsoft model)
echo   5. mistral   - 7B params (larger, more capable)
echo   6. llama2    - 7B params (Meta's Llama-2)
echo   7. codellama - 7B params (specialized for coding)
echo.
echo Or enter any other model name to search Hugging Face Hub.
echo.
set /p model="Enter model name to download (e.g., qwen): "
if "%model%"=="" goto MENU
echo.
echo 📥 Downloading model: %model%
echo 📊 Watch the progress bar for real-time download stats!
echo.
"%~dp0colossus.exe" models pull %model%
echo.
pause
goto MENU

:START_CHAT
echo.
echo 💬 Available models for chat:
echo   Use any model you've downloaded (qwen, tinyllama, gemma, phi, etc.)
echo.
set /p model="Enter model name for chat (e.g., qwen): "
if "%model%"=="" goto MENU
echo 💬 Starting chat with: %model%
echo (Type 'exit' or 'quit' to end the chat session)
echo.
"%~dp0colossus.exe" chat %model%
echo.
pause
goto MENU

:EXIT
echo.
echo 👋 Goodbye!
exit /b 0
