package email

import (
	"bytes"
	"html/template"
)

const emailTpl = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            background-color: #fafafa;
            color: #18181b;
            margin: 0;
            padding: 40px 20px;
        }
        .container {
            max-width: 500px;
            margin: 0 auto;
            background-color: #ffffff;
            border: 1px solid #e4e4e7;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
        }
        .header {
            padding: 30px 40px;
            border-bottom: 1px solid #e4e4e7;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 700;
            letter-spacing: -0.5px;
            color: #09090b;
        }
        .content {
            padding: 40px;
            font-size: 16px;
            line-height: 1.6;
            color: #3f3f46;
        }
        .btn-wrapper {
            text-align: center;
            margin: 35px 0 10px 0;
        }
        .btn {
            display: inline-block;
            background-color: #09090b;
            color: #ffffff !important;
            text-decoration: none;
            padding: 14px 28px;
            border-radius: 6px;
            font-weight: 600;
            font-size: 15px;
        }
        .footer {
            padding: 24px 40px;
            background-color: #f4f4f5;
            text-align: center;
            font-size: 13px;
            color: #71717a;
            border-top: 1px solid #e4e4e7;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>ResumeRanker</h1>
        </div>
        <div class="content">
            <h2 style="margin-top: 0; font-size: 18px; color: #18181b;">{{.Title}}</h2>
            <div style="margin-bottom: 20px;">
                {{.Message}}
            </div>
            {{if .BtnText}}
            <div class="btn-wrapper">
                <a href="{{.BtnLink}}" class="btn">{{.BtnText}}</a>
            </div>
            {{end}}
        </div>
        <div class="footer">
            <div style="margin-bottom: 15px;">
                <a href="{{.Domain}}/legal/terms.html" style="color: #71717a; text-decoration: underline; margin-right: 15px;">Terms of Service</a>
                <a href="{{.Domain}}/legal/privacy.html" style="color: #71717a; text-decoration: underline;">Privacy Policy</a>
            </div>
            {{if .SupportEmail}}
            <div style="margin-bottom: 15px;">
                Need help? <a href="mailto:{{.SupportEmail}}" style="color: #71717a; text-decoration: underline;">Contact Support</a>
            </div>
            {{end}}
            {{if .FooterNote}}
            <div style="margin-bottom: 15px;">
                {{.FooterNote}}
            </div>
            {{end}}
            &copy; 2026 ResumeRanker. All rights reserved.
        </div>
    </div>
</body>
</html>
`

var parsedTpl = template.Must(template.New("email").Parse(emailTpl))

type HTMLTemplateParams struct {
	Title        string
	Message      string
	BtnText      string
	BtnLink      string
	SupportEmail string
	Domain  string
	FooterNote   string
}

// BuildHTMLTemplate generates a premium HTML email payload.
// Message can contain simple HTML like <p> or <strong>.
func BuildHTMLTemplate(p HTMLTemplateParams) string {
	data := struct {
		Title        string
		Message      template.HTML
		BtnText      string
		BtnLink      string
		SupportEmail string
		Domain  string
		FooterNote   string
	}{
		Title:        p.Title,
		Message:      template.HTML(p.Message),
		BtnText:      p.BtnText,
		BtnLink:      p.BtnLink,
		SupportEmail: p.SupportEmail,
		Domain:  p.Domain,
		FooterNote:   p.FooterNote,
	}

	var buf bytes.Buffer
	if err := parsedTpl.Execute(&buf, data); err != nil {
		return p.Message
	}
	return buf.String()
}
