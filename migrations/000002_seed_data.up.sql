-- Core App Migration 0002: Seed initial data
-- Multi-tenant white-label SaaS brand management

-- Insert Faro agency
INSERT INTO agencies (id, name, slug, status) VALUES 
('550e8400-e29b-41d4-a716-446655440000', 'Faro', 'faro', 'active')
ON CONFLICT (slug) DO NOTHING;

-- Insert Faro branding
INSERT INTO branding (
    agency_id,
    domain,
    verified_at,
    logo_url,
    favicon_url,
    primary_color,
    secondary_color,
    theme_json,
    updated_at
) VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    'thefaro.co',
    NOW(),
    'https://example.com/logo.png',
    'https://example.com/favicon.ico',
    '#3b82f6',
    '#6b7280',
    '{
        "colors": {
            "primary": "#3b82f6",
            "secondary": "#6b7280",
            "background": "#ffffff",
            "surface": "#f9fafb",
            "text": "#111827"
        },
        "typography": {
            "fontFamily": "Inter, system-ui, sans-serif",
            "fontSize": {
                "sm": "0.875rem",
                "base": "1rem",
                "lg": "1.125rem",
                "xl": "1.25rem"
            }
        },
        "spacing": {
            "xs": "0.25rem",
            "sm": "0.5rem",
            "md": "1rem",
            "lg": "1.5rem",
            "xl": "2rem"
        }
    }'::jsonb,
    NOW()
) ON CONFLICT (agency_id) DO NOTHING;

