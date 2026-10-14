package services

import (
	"encoding/json"
	"net/http"
	"time"
)

type CryptoPrices struct {
	TonUSD  float64
	TrxUSD  float64
	GramUSD float64
}

// GetLiveCryptoPrices استعلام قیمت لحظه‌ای تون، ترون و گرام از API عمومی با فال‌بک امن
func GetLiveCryptoPrices() *CryptoPrices {
	// مقادیر پیش‌فرض (Fallback) در صورت عدم دسترسی سرور به اینترنت بین‌الملل
	prices := &CryptoPrices{
		TonUSD:  5.20,
		TrxUSD:  0.16,
		GramUSD: 0.0035,
	}

	client := http.Client{Timeout: 5 * time.Second}
	url := "https://api.coingecko.com/api/v3/simple/price?ids=the-open-network,tron,gram&vs_currencies=usd"

	resp, err := client.Get(url)
	if err != nil {
		return prices
	}
	defer resp.Body.Close()

	var data map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
		if v, ok := data["the-open-network"]["usd"]; ok && v > 0 {
			prices.TonUSD = v
		}
		if v, ok := data["tron"]["usd"]; ok && v > 0 {
			prices.TrxUSD = v
		}
		if v, ok := data["gram"]["usd"]; ok && v > 0 {
			prices.GramUSD = v
		}
	}

	return prices
}
