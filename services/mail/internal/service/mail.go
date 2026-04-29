package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Mailer interface {
	SendVerificationEmail(ctx context.Context, toEmail, username, verificationLink string) error
	SendDocumentShareEmail(ctx context.Context, toEmail, documentTitle, documentLink, role string) error
}

type resendClient struct {
	apiBase   string
	apiKey    string
	fromEmail string
	fromName  string
	client    *http.Client
}

type resendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text"`
}

func NewMailer(apiBase, apiKey, fromEmail, fromName string, client *http.Client) Mailer {
	return &resendClient{
		apiBase:   strings.TrimRight(apiBase, "/"),
		apiKey:    apiKey,
		fromEmail: fromEmail,
		fromName:  fromName,
		client:    client,
	}
}

func (c *resendClient) SendVerificationEmail(ctx context.Context, toEmail, username, verificationLink string) error {
	html := buildVerificationHTML(username, verificationLink)
	text := fmt.Sprintf("Verify your email by opening this link: %s", verificationLink)
	return c.sendEmail(ctx, toEmail, "Verify your email", html, text)
}

func (c *resendClient) SendDocumentShareEmail(ctx context.Context, toEmail, documentTitle, documentLink, role string) error {
	title := strings.TrimSpace(documentTitle)
	if title == "" {
		title = "Untitled"
	}
	cleanRole := strings.TrimSpace(role)
	if cleanRole == "" {
		cleanRole = "viewer"
	}

	html := buildDocumentShareHTML(title, documentLink, cleanRole)
	text := fmt.Sprintf(`A document "%s" was shared with you as %s. Open it here: %s`, title, cleanRole, documentLink)
	return c.sendEmail(ctx, toEmail, "A document was shared with you", html, text)
}

func (c *resendClient) sendEmail(ctx context.Context, toEmail, subject, html, text string) error {
	payload := resendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", c.fromName, c.fromEmail),
		To:      []string{toEmail},
		Subject: subject,
		HTML:    html,
		Text:    text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal resend payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBase+"/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create resend request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("call resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		resBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
		return fmt.Errorf("resend status %d: %s", resp.StatusCode, strings.TrimSpace(string(resBody)))
	}
	return nil
}

func buildVerificationHTML(username, link string) string {
	name := strings.TrimSpace(username)
	if name == "" {
		name = "there"
	}
	return fmt.Sprintf(`<p>Hi %s,</p><p>Welcome. Please verify your email by clicking the link below:</p><p><a href="%s">Verify Email</a></p>`, name, link)
}

func buildDocumentShareHTML(title, link, role string) string {
	return fmt.Sprintf(
		`<p>A document was shared with you.</p><p><strong>%s</strong></p><p>Access: %s</p><p><a href="%s">Open document</a></p>`,
		title,
		role,
		link,
	)
}
