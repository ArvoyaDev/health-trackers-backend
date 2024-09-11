package openai

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type SelectedTracker struct {
	MedicalType string       `json:"medical_type"`
	Logs        []SymptomLog `json:"logs"`
}

type SymptomLog struct {
	LogTime  string `json:"log_time"`
	Notes    string `json:"notes"`
	Severity string `json:"severity"`
	Symptoms string `json:"symptoms"`
}

func Openaimain(medicalType string, logs []SymptomLog) (string, error) {
	client := openai.NewClient(os.Getenv("OPENAI_API"))

	prompt := buildPrompt(medicalType, logs)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

// Helper function to create a dynamic prompt
func buildPrompt(medicalType string, logs []SymptomLog) string {
	// Start building the prompt
	prompt := fmt.Sprintf(
		"You are a professional pattern tracker. Based on the following symptom logs and using knowledge from %s, provide insights and recommendations for me to discuss with my primary healthcare provider.\n\n",
		medicalType,
	)

	// Add individual logs to the prompt
	for i, log := range logs {
		// Preprocess and structure the notes
		structuredNotes := preprocessNotes(log.Notes)
		prompt += fmt.Sprintf(
			"Log #%d:\n- Time: %s\n- Severity: %s\n- Symptoms: %v\n- Notes: %s\n- Structured Notes: %s\n\n",
			i+1,
			log.LogTime,
			log.Severity,
			log.Symptoms,
			log.Notes,
			structuredNotes, // Adds structured notes to the prompt
		)
	}

	// Add medical-specific recommendations
	prompt += getMedicalRecommendations(medicalType)

	// Final instruction to the AI with a strict format
	prompt += "\nThe response **must** be structured exactly as follows and should not include any other tags like <div>:\n"
	prompt += "- Patterns observed:\n  - [Patterns]\n"
	prompt += "- Recommendations:\n  - [Recommendations]\n"
	prompt += "- Holistic Techniques:\n  - [Techniques]\n"
	prompt += "\nMake sure the response includes these sections as bullet points only, with no additional HTML tags. "
	prompt += "Each section should have one or more items formatted with a hyphen (-) and text.\n"

	// Add final advisory paragraph
	prompt += "\nEnd the response with the following text in a new paragraph:\n\n"
	prompt += "It's always advisable to discuss these observations and recommendations with your healthcare provider before adopting any changes. Your healthcare provider can best guide you based on your constitution and specific health circumstances.\n\n"

	return prompt
}

func getMedicalRecommendations(medicalType string) string {
	var recommendations string

	switch strings.ToLower(medicalType) {
	case "ayurveda":
		recommendations = `
Ayurveda emphasizes balance between the body’s three doshas: Vata, Pitta, and Kapha. 
- Consider how the logs reflect imbalances in these doshas (e.g., spicy food may aggravate Pitta, axiety for Vata, depression for Kapha).
- Recommendations may include dietary changes such as avoiding hot and spicy foods for Pitta, grounding exercises for Vata, and avoiding heavy, cold foods for Kapha. Keep it related to their logs.
- Ayurvedic herbs relevant to the symptoms may be suggested, such as ashwagandha for stress or triphala for digestion.
- Consider the 20 quality gunas, the time of day, and the season for additional correlations.
`
	case "naturopathy":
		recommendations = `
Naturopathy focuses on the body's natural ability to heal.
- Consider if the logs indicate lifestyle or dietary choices that might be contributing to the symptoms.
- Recommendations may include emphasizing whole, organic foods, detoxifying the body, and reducing exposure to toxins or stress.
- Supplements relevant to the symptoms, such as probiotics for gut health or vitamin D for mood, may be suggested.
`
	case "traditional chinese medicine":
		recommendations = `
Traditional Chinese Medicine (TCM) emphasizes balance in Qi (energy flow) and the role of yin and yang.
- Look for patterns that may indicate Qi stagnation, heat, or dampness in the logs (e.g., feeling sluggish might indicate dampness, while hot weather or spicy food could indicate excess heat).
- Recommendations may include dietary modifications like cooling foods (e.g., cucumber, mint) to balance excess heat, or warming herbs for cold conditions.
- Acupuncture or herbal remedies like ginseng or licorice root may also be suggested.
`
	default:
		recommendations = `
No specific medical approach was selected. Provide general wellness advice such as maintaining a balanced diet, regular physical activity, and stress management techniques like mindfulness or yoga.
`
	}

	return recommendations
}

func preprocessNotes(note string) string {
	foodKeywords := []string{
		"food",
		"ate",
		"meal",
		"dinner",
		"lunch",
		"breakfast",
		"spicy",
		"fried",
		"sweet",
		"dairy",
		"coffee",
		"alcohol",
		"sour",
		"bitter",
		"salty",
		"microwave",
		"oven",
		"fried",
		"takeout",
		"fast",
		"snack",
		"processed",
	}
	activityKeywords := []string{
		"exercise",
		"sleep",
		"rest",
		"stress",
		"walking",
		"lying down",
		"work",
		"study",
		"meditation",
		"yoga",
		"running",
		"swimming",
		"cycling",
		"weightlifting",
		"dancing",
		"hiking",
		"climbing",
		"stretching",
		"pilates",
		"aerobics",
		"zumba",
	}
	environmentKeywords := []string{
		"weather",
		"humidity",
		"temperature",
		"hot",
		"cold",
		"rainy",
		"dry",
		"pollen",
		"dust",
		"smoke",
		"allergen",
		"mold",
		"pet",
		"animal",
		"insect",
		"pest",
		"chemical",
		"cleaning",
	}

	// Create a map of categories to match against notes
	categoryKeywords := map[string][]string{
		"Food":        foodKeywords,
		"Activity":    activityKeywords,
		"Environment": environmentKeywords,
	}

	// Track matched keywords in the note
	matchedCategories := map[string][]string{}

	// Search for matches in the note text
	for category, keywords := range categoryKeywords {
		for _, keyword := range keywords {
			if containsWord(note, keyword) {
				matchedCategories[category] = append(matchedCategories[category], keyword)
			}
		}
	}

	structuredSummary := "Detected Categories:\n"
	for category, matches := range matchedCategories {
		structuredSummary += fmt.Sprintf("- %s: %v\n", category, matches)
	}

	return structuredSummary
}

// Helper function to check if a word is present in a string
func containsWord(note string, word string) bool {
	return strings.Contains(strings.ToLower(note), strings.ToLower(word))
}
