package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"google.golang.org/genai"
)

// In-memory rate limiter per session/IP
type rateLimiter struct {
	mu      sync.Mutex
	history map[string][]time.Time
	limit   int           // max requests
	window  time.Duration // time window
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		history: make(map[string][]time.Time),
		limit:   limit,
		window:  window,
	}
	// Periodic cleanup of expired entries
	go func() {
		for range time.Tick(2 * time.Minute) {
			rl.mu.Lock()
			now := time.Now()
			for key, timestamps := range rl.history {
				var valid []time.Time
				for _, t := range timestamps {
					if now.Sub(t) < window {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(rl.history, key)
				} else {
					rl.history[key] = valid
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *rateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	var valid []time.Time
	for _, t := range rl.history[key] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		return false
	}

	valid = append(valid, now)
	rl.history[key] = valid
	return true
}

type aiService struct {
	client  *genai.Client
	model   string
	limiter *rateLimiter
}

func NewAIService(ctx context.Context, apiKey, model string) (domain.AIService, error) {
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY not set")
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("genai client: %w", err)
	}
	if model == "" {
		model = "gemini-2.0-flash"
	}
	return &aiService{
		client:  client,
		model:   model,
		limiter: newRateLimiter(10, 1*time.Minute), // Max 10 requests per minute per user/session
	}, nil
}

// CustomerChat provides customer service responses for the customer portal with strict guardrails.
func (s *aiService) CustomerChat(userID, sessionID, message string) (string, error) {
	// Apply rate limiting per user/session
	rateKey := userID
	if rateKey == "" || rateKey == "guest" {
		rateKey = sessionID
	}
	if rateKey == "" {
		rateKey = "guest-default"
	}

	if !s.limiter.Allow(rateKey) {
		return "⚠️ Had interaksi dicapai. Sila tunggu 1 minit sebelum menghantar soalan baru. / Interaction limit reached. Please wait 1 minute. / 已达到提问频率限制，请等待1分钟后再试。", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	systemPrompt := `You are the official "Asian Mum AI Nutri-Advisor" (Penasihat Nutrisi AI Asian Mum / 亚洲妈妈AI营养顾问).
Asian Mum is a premier Malaysian Halal Plant-Based Health Food and MLM Entrepreneur network.

PRODUCTS IN ASIAN MUM CATALOG:
1. Asian Mum Prebiotik Ubi Ungu (Purple Sweet Potato): 10.4g Japanese Inulin prebiotic fibre, anthocyanins for gut microbiome, smooth bowel movements, immunity.
2. Asian Mum Prebiotik Avokado & Moringa Superfood: Vitamin E, Selenium, healthy fats, anti-inflammatory, sustained vitality.
3. Asian Mum Beaufica Kranberi & Kacip Fatimah: Premium botanical sachet with genuine Kacip Fatimah, wild cranberries, pomegranate for female hormonal balance, urinary tract (UTI) health, glowing skin.
4. Asian Mum Prebiotik Mangga & Markisa: Active Vitamin C, tropical digestive relief, dopamine morning energy.
5. Asian Mum Prebiotik Koko Gelap & Gandum Oat: Dark cocoa, oat beta-glucan, low GI, satiety, heart health.
6. Asian Mum MLM Entrepreneur Program: 30-40% wholesale margin, 3-tier commission (5%, 3%, 1%), BV/PV points to e-wallet.

CRITICAL SAFETY & SCOPE RULES:
1. STRICT MEDICAL DISCLAIMER: Asian Mum products are functional food supplements and NOT pharmaceuticals/medicine. Strictly NEVER give medical diagnoses, prescribe treatments, or claim to cure diseases.
   ALWAYS include a clear medical disclaimer if the user mentions illnesses, diseases, severe pain, medication, or medical symptoms:
   "Peringatan: Produk Asian Mum adalah makanan kesihatan tambahan dan bukan ubat. Sila dapatkan nasihat daripada doktor atau pakar perubatan bertauliah untuk sebarang diagnosis atau rawatan perubatan."
2. STRICT SCOPE REFUSAL: Do NOT answer questions unrelated to Asian Mum, nutrition, gut health, botanical wellness, or the Asian Mum business. If the user asks about coding, politics, homework, celebrities, games, or general trivia, politely decline:
   "Maaf, saya hanya boleh membantu soalan berkaitan makanan kesihatan Asian Mum, khasiat prebiotik & botani, serta program usahawan kami."
3. Multilingual: Reply politely in the language the customer used (Bahasa Melayu, English, or 简体中文). Keep answers concise, helpful, and welcoming.`

	result, err := s.client.Models.GenerateContent(
		ctx,
		s.model,
		genai.Text(fmt.Sprintf("%s\n\nCustomer message: %s", systemPrompt, message)),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("AI service error: %w", err)
	}
	if len(result.Candidates) == 0 {
		return "", errors.New("no response from AI")
	}
	return result.Text(), nil
}

// AdminAnalytics provides AI-powered analytics insights for admins.
func (s *aiService) AdminAnalytics(adminID, query string, dbContext string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	systemPrompt := fmt.Sprintf(`You are an intelligent business analytics assistant for Asian Mum, a Malaysian health food & MLM e-commerce platform.
You have access to the following business data context:
%s

Provide insightful, data-driven answers to help administrators make better business decisions.
Focus on actionable recommendations, product sales, commission trends, and KPI analysis.
Format responses clearly with bullet points or sections when appropriate.`, dbContext)

	result, err := s.client.Models.GenerateContent(
		ctx,
		s.model,
		genai.Text(fmt.Sprintf("%s\n\nAdmin query: %s", systemPrompt, query)),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("AI service error: %w", err)
	}
	if len(result.Candidates) == 0 {
		return "", errors.New("no response from AI")
	}
	return result.Text(), nil
}

// NoopAIService is used when Gemini API key is not configured.
type noopAIService struct{}

func NewNoopAIService() domain.AIService { return &noopAIService{} }

func (n *noopAIService) CustomerChat(_, _, _ string) (string, error) {
	return "⚠️ Penasihat AI Nutri-Advisor sedang dikemaskini. Sila hubungi khidmat pelanggan kami atau rujuk doktor bertauliah untuk nasihat perubatan.", nil
}

func (n *noopAIService) AdminAnalytics(_, _, _ string) (string, error) {
	return "AI analytics is not configured. Please set the GEMINI_API_KEY environment variable.", nil
}
