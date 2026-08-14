package email

import (
	"strings"
	"testing"
)

func TestBuildHTMLTemplate(t *testing.T) {
	tests := []struct {
		name     string
		params   HTMLTemplateParams
		contains []string
		excludes []string
	}{
		{
			name: "renders required fields",
			params: HTMLTemplateParams{
				Title:   "Verify your email",
				Message: "Please verify your account.",
				Domain:  "https://example.com",
			},
			contains: []string{
				"Verify your email",
				"Please verify your account.",
				"https://example.com/legal/terms.html",
				"https://example.com/legal/privacy.html",
			},
		},
		{
			name: "renders button when configured",
			params: HTMLTemplateParams{
				Title:   "Verify your email",
				Message: "Please verify your account.",
				BtnText: "Verify Email",
				BtnLink: "https://example.com/verify?token=abc",
				Domain:  "https://example.com",
			},
			contains: []string{
				"Verify Email",
				`href="https://example.com/verify?token=abc"`,
			},
		},
		{
			name: "omits button when not configured",
			params: HTMLTemplateParams{
				Title:   "Notice",
				Message: "Your account was updated.",
				Domain:  "https://example.com",
			},
			excludes: []string{
				"Verify Email",
				`href="https://example.com/verify`,
			},
		},
		{
			name: "renders support email when configured",
			params: HTMLTemplateParams{
				Title:        "Notice",
				Message:      "Your account was updated.",
				SupportEmail: "support@example.com",
				Domain:       "https://example.com",
			},
			contains: []string{
				"Need help?",
				"mailto:support@example.com",
				"support@example.com",
			},
		},
		{
			name: "renders footer note when configured",
			params: HTMLTemplateParams{
				Title:      "Notice",
				Message:    "Your account was updated.",
				FooterNote: "This is an automated message.",
				Domain:     "https://example.com",
			},
			contains: []string{
				"This is an automated message.",
			},
		},
		{
			name: "preserves trusted message html",
			params: HTMLTemplateParams{
				Title:   "Notice",
				Message: "<p>Hello <strong>Kishan</strong>.</p>",
				Domain:  "https://example.com",
			},
			contains: []string{
				"<p>Hello <strong>Kishan</strong>.</p>",
			},
		},
		{
			name: "escapes ordinary template values",
			params: HTMLTemplateParams{
				Title:   `<script>alert("xss")</script>`,
				Message: "Hello",
				BtnText: `<img src=x onerror=alert(1)>`,
				BtnLink: `https://example.com/?q="test"`,
				Domain:  "https://example.com",
			},
			contains: []string{
				"&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;",
				"&lt;img src=x onerror=alert(1)&gt;",
			},
			excludes: []string{
				`<script>alert("xss")</script>`,
				`<img src=x onerror=alert(1)>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildHTMLTemplate(tt.params)

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("expected rendered template to contain %q", want)
				}
			}

			for _, unwanted := range tt.excludes {
				if strings.Contains(got, unwanted) {
					t.Errorf("expected rendered template not to contain %q", unwanted)
				}
			}
		})
	}
}
