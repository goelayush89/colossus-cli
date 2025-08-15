@echo off
REM Quick Start - Colossus Server with Enhanced Features
echo ========================================
echo    🚀 Colossus CLI Server v1.1
echo ========================================
echo.
echo ✨ New Features:
echo   📊 Visual progress bars for downloads
echo   🚀 Enhanced GGUF model repositories
echo   ⚡ 7 guaranteed downloadable models
echo.
echo 🌐 Server will be available at:
echo   - API: http://localhost:11434
echo   - Health: http://localhost:11434/
echo.
echo 💡 Quick commands after server starts:
echo   - Download model: colossus.exe models pull qwen
echo   - Start chat: colossus.exe chat qwen
echo   - List models: colossus.exe models list
echo.
echo Press Ctrl+C to stop the server
echo.

REM Start the server
"%~dp0colossus.exe" serve --verbose

echo.
echo 🛑 Server stopped.
pause
