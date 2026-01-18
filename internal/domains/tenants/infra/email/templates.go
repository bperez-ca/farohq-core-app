package email

import (
	"fmt"
	"strings"
	"time"
)

// InviteEmailData contains all data needed for invitation email templates
type InviteEmailData struct {
	// Invite information
	InviteEmail string
	RoleName    string
	InviteURL   string
	ExpiresAt   time.Time

	// Agency/Tenant information
	AgencyName string
	Tier       string // "starter", "growth", "scale"

	// Branding information
	LogoURL        string
	PrimaryColor   string
	SecondaryColor string
	HidePoweredBy  bool // true for white-label (Growth+), false for gray-label

	// User information (optional - may not be available)
	InviteeFirstName string // Extracted from email if not available
	InviterName      string // Name of person who sent invite
	InviterEmail     string // Email of person who sent invite
}

// GetRoleDisplayName returns a human-readable role name
func GetRoleDisplayName(role string) string {
	switch strings.ToLower(role) {
	case "owner":
		return "Owner"
	case "admin":
		return "Admin"
	case "staff":
		return "Staff"
	case "viewer":
		return "Viewer"
	case "clientviewer":
		return "Client Viewer"
	default:
		return strings.Title(role)
	}
}

// ExtractFirstName extracts first name from email (fallback)
func ExtractFirstName(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		localPart := parts[0]
		// Try to extract name from email like "john.doe@example.com" -> "John"
		nameParts := strings.Split(localPart, ".")
		if len(nameParts) > 0 {
			return strings.Title(nameParts[0])
		}
		return strings.Title(localPart)
	}
	return ""
}

// BuildInviteEmailSubject builds the email subject based on branding mode
func BuildInviteEmailSubject(data InviteEmailData) string {
	if data.HidePoweredBy {
		// White-label (Growth/Scale)
		return fmt.Sprintf("You're invited to join %s", data.AgencyName)
	} else if data.Tier == "starter" {
		// Gray-label (Starter)
		return fmt.Sprintf("You're invited to join %s on FARO HQ", data.AgencyName)
	} else {
		// FARO-branded (fallback)
		return "You're invited to join a workspace on FARO HQ"
	}
}

// BuildInviteEmailFromName builds the "From" name based on branding mode
func BuildInviteEmailFromName(data InviteEmailData) string {
	if data.HidePoweredBy {
		// White-label (Growth/Scale)
		return data.AgencyName
	} else if data.Tier == "starter" {
		// Gray-label (Starter)
		return fmt.Sprintf("%s via FARO HQ", data.AgencyName)
	} else {
		// FARO-branded (fallback)
		return "FARO HQ"
	}
}

// BuildInviteEmailHTML builds the HTML email body with branding
func BuildInviteEmailHTML(data InviteEmailData) (string, error) {
	// Determine branding mode
	isWhiteLabel := data.HidePoweredBy && (data.Tier == "growth" || data.Tier == "scale")
	isGrayLabel := data.Tier == "starter"
	
	// Get invitee name (use first name if available, otherwise extract from email)
	inviteeName := data.InviteeFirstName
	if inviteeName == "" {
		inviteeName = ExtractFirstName(data.InviteEmail)
	}
	if inviteeName == "" {
		inviteeName = "there"
	}

	// Get inviter name
	inviterName := data.InviterName
	if inviterName == "" {
		inviterName = ExtractFirstName(data.InviterEmail)
	}
	if inviterName == "" {
		inviterName = "A team member"
	}

	// Get role display name
	roleName := GetRoleDisplayName(data.RoleName)

	// Format expiration date
	expiresAtFormatted := data.ExpiresAt.Format("January 2, 2006 at 3:04 PM MST")

	// Determine button color (use primary color if available, otherwise default)
	buttonColor := data.PrimaryColor
	if buttonColor == "" {
		buttonColor = "#007bff"
	}

	// Build logo HTML if available and white-label
	logoHTML := ""
	if isWhiteLabel && data.LogoURL != "" {
		logoHTML = fmt.Sprintf(`<div style="text-align: center; margin-bottom: 30px;">
			<img src="%s" alt="%s" style="max-width: 200px; max-height: 60px; height: auto;" />
		</div>`, data.LogoURL, data.AgencyName)
	}

	// Build footer based on branding mode
	var footerHTML string
	if isWhiteLabel {
		// White-label: Just agency name, no FARO branding
		footerHTML = fmt.Sprintf(`<div style="margin-top: 40px; padding-top: 20px; border-top: 1px solid #e0e0e0; text-align: center; color: #666; font-size: 12px;">
			<p>%s Team</p>
		</div>`, data.AgencyName)
	} else if isGrayLabel {
		// Gray-label: Agency name + "Powered by FARO HQ"
		footerHTML = fmt.Sprintf(`<div style="margin-top: 40px; padding-top: 20px; border-top: 1px solid #e0e0e0; text-align: center; color: #666; font-size: 12px;">
			<p><strong>%s</strong></p>
			<p>Workspace powered by FARO HQ</p>
		</div>`, data.AgencyName)
	} else {
		// FARO-branded: FARO branding
		footerHTML = `<div style="margin-top: 40px; padding-top: 20px; border-top: 1px solid #e0e0e0; text-align: center; color: #666; font-size: 12px;">
			<p><strong>FARO HQ</strong></p>
			<p>Local Visibility Management Platform</p>
		</div>`
	}

	// Build greeting based on whether we have a name
	greeting := "Hi " + inviteeName + ","
	if inviteeName == "there" {
		greeting = "Hello,"
	}

	// Build main content based on branding mode
	var mainContent string
	if isWhiteLabel {
		// White-label version
		mainContent = fmt.Sprintf(`<p>%s has invited you to join %s as <strong>%s</strong> in our Local Visibility HQ.</p>`, inviterName, data.AgencyName, roleName)
	} else if isGrayLabel {
		// Gray-label version
		mainContent = fmt.Sprintf(`<p>%s has invited you to join %s as <strong>%s</strong> inside their FARO HQ workspace.</p>`, inviterName, data.AgencyName, roleName)
	} else {
		// FARO-branded version
		mainContent = fmt.Sprintf(`<p>%s has invited you to join %s as <strong>%s</strong> on FARO HQ.</p>`, inviterName, data.AgencyName, roleName)
	}

	// Build the full HTML template
	htmlTemplate := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<style>
		body { 
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; 
			line-height: 1.6; 
			color: #333333; 
			background-color: #f5f5f5;
			margin: 0;
			padding: 0;
		}
		.email-container {
			max-width: 600px;
			margin: 0 auto;
			background-color: #ffffff;
		}
		.content {
			padding: 40px 30px;
		}
		.logo-container {
			text-align: center;
			margin-bottom: 30px;
		}
		.logo-container img {
			max-width: 200px;
			max-height: 60px;
			height: auto;
		}
		h1 {
			color: #1a1a1a;
			font-size: 24px;
			font-weight: 600;
			margin: 0 0 20px 0;
		}
		p {
			margin: 0 0 16px 0;
			color: #4a4a4a;
			font-size: 16px;
		}
		.button-container {
			text-align: center;
			margin: 30px 0;
		}
		.button {
			display: inline-block;
			padding: 14px 32px;
			background-color: %s;
			color: #ffffff;
			text-decoration: none;
			border-radius: 6px;
			font-weight: 600;
			font-size: 16px;
			box-shadow: 0 2px 4px rgba(0,0,0,0.1);
		}
		.button:hover {
			opacity: 0.9;
		}
		.link-fallback {
			margin-top: 20px;
			padding: 15px;
			background-color: #f8f9fa;
			border-radius: 4px;
			word-break: break-all;
			font-size: 14px;
			color: #666;
		}
		.footer {
			margin-top: 40px;
			padding-top: 20px;
			border-top: 1px solid #e0e0e0;
			text-align: center;
			color: #666;
			font-size: 12px;
		}
		.expires-note {
			margin-top: 20px;
			padding: 12px;
			background-color: #fff3cd;
			border-left: 4px solid #ffc107;
			border-radius: 4px;
			font-size: 14px;
		}
	</style>
</head>
<body>
	<div class="email-container">
		<div class="content">
			%s
			<h1>You're Invited!</h1>
			%s
			%s
			<div class="button-container">
				<a href="%s" class="button">Accept Invitation</a>
			</div>
			<p>If the button doesn't work, copy and paste this link into your browser:</p>
			<div class="link-fallback">%s</div>
			<div class="expires-note">
				<strong>Note:</strong> This invitation expires on %s.
			</div>
			<p>If you weren't expecting this invitation, you can safely ignore this email.</p>
			%s
		</div>
	</div>
</body>
</html>`, buttonColor, logoHTML, greeting, mainContent, data.InviteURL, data.InviteURL, expiresAtFormatted, footerHTML)

	return htmlTemplate, nil
}

// BuildInviteEmailText builds the plain text email body
func BuildInviteEmailText(data InviteEmailData) string {
	// Determine branding mode
	isWhiteLabel := data.HidePoweredBy && (data.Tier == "growth" || data.Tier == "scale")
	isGrayLabel := data.Tier == "starter"

	// Get invitee name
	inviteeName := data.InviteeFirstName
	if inviteeName == "" {
		inviteeName = ExtractFirstName(data.InviteEmail)
	}
	if inviteeName == "" {
		inviteeName = "there"
	}

	// Get inviter name
	inviterName := data.InviterName
	if inviterName == "" {
		inviterName = ExtractFirstName(data.InviterEmail)
	}
	if inviterName == "" {
		inviterName = "A team member"
	}

	// Get role display name
	roleName := GetRoleDisplayName(data.RoleName)

	// Format expiration date
	expiresAtFormatted := data.ExpiresAt.Format("January 2, 2006 at 3:04 PM MST")

	// Build greeting
	greeting := "Hi " + inviteeName + ","
	if inviteeName == "there" {
		greeting = "Hello,"
	}

	// Build main content
	var mainContent string
	if isWhiteLabel {
		mainContent = fmt.Sprintf("%s has invited you to join %s as %s in our Local Visibility HQ.", inviterName, data.AgencyName, roleName)
	} else if isGrayLabel {
		mainContent = fmt.Sprintf("%s has invited you to join %s as %s inside their FARO HQ workspace.", inviterName, data.AgencyName, roleName)
	} else {
		mainContent = fmt.Sprintf("%s has invited you to join %s as %s on FARO HQ.", inviterName, data.AgencyName, roleName)
	}

	// Build footer
	var footer string
	if isWhiteLabel {
		footer = fmt.Sprintf("\n\nThanks,\n%s Team", data.AgencyName)
	} else if isGrayLabel {
		footer = fmt.Sprintf("\n\n—\n%s\nWorkspace powered by FARO HQ", data.AgencyName)
	} else {
		footer = "\n\n—\nFARO HQ\nLocal Visibility Management Platform"
	}

	return fmt.Sprintf(`You're Invited!

%s

%s

Click the button below to accept your invitation and set up your account:

Accept your invitation by visiting:
%s

Note: This invitation expires on %s.

If you weren't expecting this invitation, you can safely ignore this email.%s`, greeting, mainContent, data.InviteURL, expiresAtFormatted, footer)
}
