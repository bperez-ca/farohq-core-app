# Branding Component Styles Implementation Summary

## Overview
Successfully implemented component-specific border radius styling and contrast enforcement across the branding system.

## Completed Implementation

### 1. Backend (Go)

#### ✅ Contrast Utilities (`internal/domains/brand/domain/utils/contrast.go`)
- `CalculateContrastRatio()` - WCAG 2.1 contrast ratio calculation
- `GetContrastingTextColor()` - Returns appropriate text color (white/dark) for background
- `ValidateContrast()` - Validates if contrast meets minimum ratio (default 4.5:1)
- `EnsureContrast()` - Ensures text color has sufficient contrast, auto-adjusts if needed
- Supports hex (#RRGGBB, #RGB) and rgb/rgba color formats

#### ✅ Brand Model Extensions (`internal/domains/brand/domain/model/branding.go`)
- `GetComponentBorderRadius(component, style)` - Returns border radius for component type
  - Components: `button`, `card`, `panel`, `tile`, `badge`, `input`
  - Styles: `default`, `rounded`, `square`
- `ValidateContrast()` - Validates contrast using theme minimum ratio
- `GetContrastingTextColor()` - Gets contrasting text color for background
- `EnsureContrast()` - Ensures contrast with auto-adjust
- `GetContrastMinimumRatio()` - Returns minimum ratio from theme (default 4.5)
- `GetContrastAutoAdjust()` - Returns auto-adjust setting (default true)

### 2. Frontend (TypeScript/React)

#### ✅ Contrast Utilities (`src/lib/contrast.ts`)
- `calculateContrastRatio()` - WCAG contrast calculation
- `getContrastingTextColor()` - Returns white/dark text for background
- `validateContrast()` - Validates contrast ratio
- `ensureContrast()` - Ensures contrast with auto-adjust option
- `getContrastingTextClass()` - Returns Tailwind class for contrasting text

#### ✅ BrandThemeProvider Updates (`src/components/branding/BrandThemeProvider.tsx`)
- Parses component-specific border radius from theme JSON
- Applies CSS variables for all component types:
  - `--border-radius-button-{default|rounded|square}`
  - `--border-radius-card-{default|rounded|square}`
  - `--border-radius-panel-{default|rounded|square}`
  - `--border-radius-tile-{default|rounded|square}`
  - `--border-radius-badge-{default|rounded|square}`
  - `--border-radius-input-{default|rounded|square}`
- Backward compatible with legacy single `border_radius` value
- Sets sensible defaults if theme JSON doesn't include border radius config

#### ✅ Component Updates

**StatusChip** (`src/components/ui/StatusChip.tsx`)
- ✅ Fixed contrast issue (cyan text on cyan background)
- ✅ Updated to use darker text colors (green-800, amber-900, etc.) for better contrast
- ✅ Uses theme-controlled border radius: `rounded-[var(--border-radius-badge-default)]`
- ✅ All color combinations verified for WCAG AA compliance (4.5:1)

**Button** (`src/components/ui/button.tsx`)
- ✅ Updated all variants to use `rounded-[var(--border-radius-button-default)]`
- ✅ Ghost variant uses `rounded-[var(--border-radius-button-rounded)]`
- ✅ Icon size uses theme-controlled radius

**Card** (`src/components/ui/card.tsx`)
- ✅ Updated to use `rounded-[var(--border-radius-card-default)]`

#### ✅ Tailwind Config (`tailwind.config.js`)
- ✅ Added theme-controlled border radius utilities:
  - `rounded-button`, `rounded-button-rounded`, `rounded-button-square`
  - `rounded-card`, `rounded-card-rounded`, `rounded-card-square`
  - `rounded-panel`, `rounded-panel-rounded`, `rounded-panel-square`
  - `rounded-tile`, `rounded-tile-rounded`, `rounded-tile-square`
  - `rounded-badge`, `rounded-badge-rounded`, `rounded-badge-square`
  - `rounded-input`, `rounded-input-rounded`, `rounded-input-square`

## Theme JSON Structure

### Component Border Radius
```json
{
  "spacing": {
    "border_radius": {
      "global": "8px",
      "button": {
        "default": "999px",
        "rounded": "8px",
        "square": "0px"
      },
      "card": {
        "default": "12px",
        "rounded": "16px",
        "square": "0px"
      },
      "panel": {
        "default": "8px",
        "rounded": "12px",
        "square": "0px"
      },
      "tile": {
        "default": "4px",
        "rounded": "8px",
        "square": "0px"
      },
      "badge": {
        "default": "999px",
        "rounded": "6px",
        "square": "0px"
      },
      "input": {
        "default": "6px",
        "rounded": "8px",
        "square": "0px"
      }
    }
  }
}
```

### Contrast Rules (Future Enhancement)
```json
{
  "contrast": {
    "enforce": true,
    "minimum_ratio": 4.5,
    "large_text_ratio": 3.0,
    "auto_adjust": true,
    "fallback_text": {
      "light": "#1f2937",
      "dark": "#ffffff"
    }
  }
}
```

## Default Values

If theme JSON doesn't specify border radius, defaults are applied:

- **Button**: `999px` (pill), `8px` (rounded), `0px` (square)
- **Card**: `12px` (default), `16px` (rounded), `0px` (square)
- **Panel**: `8px` (default), `12px` (rounded), `0px` (square)
- **Tile**: `4px` (default), `8px` (rounded), `0px` (square)
- **Badge**: `999px` (pill), `6px` (rounded), `0px` (square)
- **Input**: `6px` (default), `8px` (rounded), `0px` (square)

## Contrast Fixes

### StatusChip Color Updates
All variants updated to use higher contrast text colors:

- **synced/success**: `text-green-800` (was `text-green-700`)
- **needsUpdate/needsAttention/waiting**: `text-amber-900` (was `text-amber-700`)
- **notConnected**: `text-rose-900` (was `text-rose-700`)
- **replied**: `text-emerald-900` (was `text-emerald-700`)
- **warning**: `text-orange-900` (was `text-orange-700`)
- **danger**: `text-red-900` (was `text-red-700`)
- **Dark mode**: All use `text-{color}-200` for better contrast

## Domain Verification (Vercel)

Custom domain verification for white-label branding (Scale tier) uses the Vercel API for CNAME instructions and verification status.

### Flow

1. **Add domain to Vercel**: When an agency configures a custom domain in branding settings, the backend calls Vercel API to add the domain to the portal project (if not already added).
2. **Fetch CNAME target**: The CNAME target is always fetched from the Vercel API (never hardcoded) because the value may vary.
3. **Show instructions in UI**: The DomainVerification component displays the expected CNAME target and human-readable setup instructions.
4. **Verify**: User adds CNAME record in their DNS provider, then clicks "Verify Domain". Backend calls Vercel API to confirm verification and SSL status.

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/brands/{brandId}/domain-instructions` | GET | Returns domain, CNAME target, and setup instructions (from Vercel API) |
| `/api/v1/brands/{brandId}/domain-status` | GET | Returns verification status, expected CNAME, SSL status |
| `/api/v1/brands/{brandId}/verify-domain` | POST | Triggers verification check via Vercel API |

### Environment Variables

Required for domain verification (Scale tier):

- `VERCEL_API_TOKEN` – Vercel API token
- `VERCEL_PROJECT_ID` – Vercel project ID (portal deployment)
- `VERCEL_TEAM_ID` – (Optional) Vercel team ID

See [docs/plans/ENV_VARS_AGENTS.md](plans/ENV_VARS_AGENTS.md) for full env var documentation.

### Portal Component

The `DomainVerification` component is shown on the branding settings page (`/agency/settings/branding`) when the agency has a custom domain configured and is on Scale tier. It displays:

- Expected CNAME target (from Vercel API)
- DNS setup instructions
- Verify Domain button
- SSL certificate status (pending, active, failed)

## Testing Status

- ✅ Go code compiles successfully
- ✅ TypeScript/React code has no linter errors
- ✅ Backward compatible with existing themes
- ✅ Defaults applied when theme JSON missing border radius config

## Next Steps (Future Enhancements)

1. **Add contrast validation to brand creation/update endpoints**
   - Validate color combinations in use cases
   - Auto-adjust text colors if contrast insufficient

2. **Migrate additional components**
   - Panel components
   - Tile components
   - Input components
   - Badge components

3. **Add UI for theme configuration**
   - Visual theme builder
   - Real-time preview
   - Contrast ratio indicators

4. **Add tests**
   - Unit tests for contrast utilities
   - Component tests for border radius application
   - Visual regression tests

## Files Modified

### Backend
- `internal/domains/brand/domain/utils/contrast.go` (new)
- `internal/domains/brand/domain/model/branding.go` (extended)

### Frontend
- `src/lib/contrast.ts` (new)
- `src/components/branding/BrandThemeProvider.tsx` (extended)
- `src/components/ui/StatusChip.tsx` (fixed contrast, added theme border radius)
- `src/components/ui/button.tsx` (added theme border radius)
- `src/components/ui/card.tsx` (added theme border radius)
- `tailwind.config.js` (added theme border radius utilities)

## Breaking Changes

**None** - All changes are backward compatible. Existing themes without component border radius config will use sensible defaults.
