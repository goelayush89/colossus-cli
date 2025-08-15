@echo off
REM Quick Test - Download and Chat with Qwen (smallest model)
echo.
echo ========================================
echo    🚀 Colossus CLI v1.1 - Quick Test
echo ========================================
echo.
echo ✨ NEW: Visual progress bars for downloads!
echo.
echo This will:
echo 1. Start the server in background
echo 2. Download qwen model (0.5B - smallest) with progress bar
echo 3. Start a chat session
echo.
pause

echo 🚀 Starting Colossus server...
start /B "%~dp0colossus.exe" serve
timeout /t 3 /nobreak >nul

echo 📥 Downloading qwen model (smallest)...
"%~dp0colossus.exe" models pull qwen

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Model downloaded successfully!
    echo 💬 Starting chat session...
    echo (Type your messages and press Enter. Type 'exit' to quit)
    echo.
    "%~dp0colossus.exe" chat qwen
) else (
    echo.
    echo ❌ Model download failed. Check your internet connection.
    pause
)

echo.
echo 🛑 Stopping server...
taskkill /F /IM colossus.exe >nul 2>&1
echo 👋 Test complete!
pause
