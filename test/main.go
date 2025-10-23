package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	userID := "userName"
	password := "Password"
	totpSecret := os.Args[1]
	fmt.Printf("TOTP Secret: '%s'\n", totpSecret)

	if userID == "" || password == "" || totpSecret == "" {
		log.Fatal("Please set ZERODHA_USER, ZERODHA_PASS, ZERODHA_TOTP_SECRET env vars")
	}

	// Launch Chrome in non-headless mode (set Headless(true) for background automation)
	url := launcher.New().Headless(false).MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()
	defer browser.MustClose()

	// Navigate to login page
	page := browser.MustPage("https://kite.zerodha.com/")

	// Fill user ID & password
	page.MustElement(`#userid`).MustInput(userID)
	page.MustElement(`#password`).MustInput(password)
	page.MustElement(`button[type=submit]`).MustClick()

	// Wait for 2FA screen
	page.MustWaitLoad()

	// Enter TOTP
	// page.MustElement(`#userid`).MustInput(totpSecret)
	page.MustElement(`input[type="number"][maxlength="6"]`).MustInput(totpSecret)

	// Click Continue
	continueBtn := page.MustElementR("button", "Continue|Login")
	continueBtn.MustClick()
	// continueBtn.MustWaitVisible().MustClick()  // Wait until dashboard loads
	page.MustWaitLoad()

	fmt.Println("Login successful!")

	// Try to find consent modal ("I understand" button)

	btns := page.MustElements("button")
	if len(btns) > 0 {
		fmt.Println("⚠️ Consent modal detected, clicking 'I understand'...")
		btns[0].MustClick()
		page.MustWaitLoad()
	} else {
		fmt.Println("✅ No consent modal detected")
	}

	// Example: Go to Holdings

	// Capture Holdings JSON from API (optional: hijack requests)
	// For now, just print page title
	// Get full HTML of the current page

	page.MustNavigate("https://console.zerodha.com/dashboard")
	page.MustWaitLoad()
	html := page.MustEval(`() => document.documentElement.outerHTML`).String()
	fmt.Println("========= PAGE HTML =========")
	fmt.Println(html)
	fmt.Println("=============================")
	// Keep browser open for inspection

	time.Sleep(5 * time.Second)
}
