# Branding Component Styles & Contrast Enforcement Plan

## Overview
This plan extends the branding system to support component-specific styling (border radius, corners) and enforces contrast rules to ensure text is always readable against its background color.

## Current State Analysis

### Existing Branding System
- **Backend**: `theme_json` stored as `map[string]interface{}` in `branding` table
- **Frontend**: `BrandThemeProvider` applies CSS variables from theme
- **Current Support**:
  - Colors: `brand`, `brand_hover`, `accent`, `background`, `foreground`
  - Typography: `font_family`, `font_size_base`, `line_height_base`
  - Spacing: `border_radius` (single global value)

### Issues Identified
1. **Inconsistent Border Radius**: Components use hardcoded values (`rounded-lg`, `rounded-xl`, `rounded-full`, `rounded-md`) instead of brand-controlled values
2. **No Component-Specific Styling**: All components share the same border radius, but different component types (buttons, cards, panels, tiles) should have distinct styles
3. **Contrast Violations**: Text color can match background color (e.g., cyan text on cyan background), making content illegible
4. **No Validation**: No server-side or client-side validation ensures contrast ratios meet WCAG standards

## Proposed Solution

### Phase 1: Extend Theme JSON Schema

#### 1.1 Component-Specific Border Radius Structure
Extend `theme_json.spacing` to include component-specific border radius:

```json
{
  "spacing": {
    "border_radius": {
      "global": "8px",           // Default fallback
      "button": {
        "default": "999px",       // Fully rounded (pill shape)
        "rounded": "8px",         // Rounded corners
        "square": "0px"           // Straight corners
      },
      "card": {
        "default": "12px",        // Rounded corners
        "rounded": "16px",        // More rounded
        "square": "0px"           // Straight corners
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
        "default": "999px",       // Pill shape
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

#### 1.2 Contrast Rules Structure
Add contrast validation rules to `theme_json`:

```json
{
  "contrast": {
    "enforce": true,              // Enable/disable contrast enforcement
    "minimum_ratio": 4.5,         // WCAG AA standard (4.5:1 for normal text)
    "large_text_ratio": 3.0,      // WCAG AA for large text (18pt+ or 14pt+ bold)
    "auto_adjust": true,          // Automatically adjust text color if contrast fails
    "fallback_text": {
      "light": "#1f2937",         // Dark text for light backgrounds
      "dark": "#ffffff"           // White text for dark backgrounds
    }
  }
}
```

### Phase 2: Backend Implementation

#### 2.1 Update Brand Model (Go)
**File**: `internal/domains/brand/domain/model/branding.go`

Add validation methods:
- `ValidateThemeJSON()` - Validates theme JSON structure
- `GetComponentBorderRadius(component, style)` - Returns border radius for component
- `EnforceContrast(backgroundColor, textColor)` - Validates and adjusts contrast

#### 2.2 Update Create/Update Brand Use Cases
**Files**: 
- `internal/domains/brand/app/usecases/create_brand.go`
- `internal/domains/brand/app/usecases/update_brand.go` (if exists)

Add validation:
- Validate theme JSON structure on create/update
- Ensure contrast rules are enforced
- Set default values if missing

#### 2.3 Add Contrast Utilities
**New File**: `internal/domains/brand/domain/utils/contrast.go`

```go
package utils

import (
    "math"
)

// CalculateContrastRatio calculates WCAG contrast ratio between two colors
func CalculateContrastRatio(color1, color2 string) float64 {
    // Implementation using relative luminance
}

// GetContrastingTextColor returns appropriate text color for background
func GetContrastingTextColor(backgroundColor string) string {
    // Returns white or dark based on background luminance
}

// ValidateContrast validates if contrast meets minimum ratio
func ValidateContrast(backgroundColor, textColor string, minRatio float64) bool {
    // Returns true if contrast is sufficient
}
```

### Phase 3: Frontend Implementation

#### 3.1 Update BrandThemeProvider
**File**: `src/components/branding/BrandThemeProvider.tsx`

**Changes**:
1. Parse component-specific border radius from theme
2. Apply CSS variables for each component type:
   ```typescript
   --border-radius-button-default: 999px;
   --border-radius-button-rounded: 8px;
   --border-radius-button-square: 0px;
   --border-radius-card-default: 12px;
   // ... etc
   ```
3. Implement contrast validation and auto-adjustment
4. Apply contrast rules to all color combinations

#### 3.2 Create Contrast Utility Functions
**New File**: `src/lib/contrast.ts`

```typescript
/**
 * Calculate WCAG contrast ratio between two colors
 */
export function calculateContrastRatio(color1: string, color2: string): number

/**
 * Get appropriate text color for background (white or dark)
 */
export function getContrastingTextColor(backgroundColor: string): string

/**
 * Validate contrast meets minimum ratio
 */
export function validateContrast(
  backgroundColor: string, 
  textColor: string, 
  minRatio: number = 4.5
): boolean

/**
 * Ensure text color has sufficient contrast, auto-adjust if needed
 */
export function ensureContrast(
  backgroundColor: string,
  textColor: string,
  options?: { minRatio?: number; autoAdjust?: boolean }
): string
```

#### 3.3 Update Component Styles

**Button Component** (`src/components/ui/button.tsx`):
- Replace hardcoded `rounded-full` with CSS variable
- Support variants: `default` (pill), `rounded`, `square`
- Apply contrast validation to text colors

**Card Component** (`src/components/ui/card.tsx`):
- Replace hardcoded `rounded-xl` with CSS variable
- Support variants based on theme

**StatusChip Component** (`src/components/ui/StatusChip.tsx`):
- Fix contrast issue (cyan text on cyan background)
- Use `ensureContrast()` for all color combinations

**All UI Components**:
- Replace hardcoded border radius classes with CSS variables
- Apply contrast validation to text/background combinations

#### 3.4 Update Tailwind Config
**File**: `tailwind.config.js`

Add dynamic border radius values from CSS variables:
```javascript
borderRadius: {
  // ... existing values
  'button': 'var(--border-radius-button-default)',
  'button-rounded': 'var(--border-radius-button-rounded)',
  'button-square': 'var(--border-radius-button-square)',
  'card': 'var(--border-radius-card-default)',
  'card-rounded': 'var(--border-radius-card-rounded)',
  'card-square': 'var(--border-radius-card-square)',
  // ... etc
}
```

### Phase 4: Component Migration

#### 4.1 Priority Components (High Impact)
1. **Button** - Most visible component, currently uses `rounded-full`
2. **Card** - Used extensively, currently uses `rounded-xl`
3. **StatusChip** - Has contrast violation (cyan on cyan)
4. **Badge** - Used for status indicators
5. **Input** - Form elements

#### 4.2 Secondary Components
1. **Panel** - Dashboard panels
2. **Tile** - Grid items
3. **Modal** - Dialog boxes
4. **Tooltip** - Hover elements
5. **Dropdown** - Menu items

#### 4.3 Migration Pattern
For each component:
1. Identify current border radius usage
2. Replace with CSS variable from theme
3. Add contrast validation for text/background colors
4. Test with different brand themes
5. Update component documentation

### Phase 5: Validation & Testing

#### 5.1 Backend Validation
- Unit tests for contrast calculation
- Integration tests for theme JSON validation
- Test edge cases (very light/dark colors)

#### 5.2 Frontend Validation
- Visual regression tests for component styles
- Contrast ratio tests for all color combinations
- Cross-browser testing
- Dark mode compatibility

#### 5.3 Accessibility Testing
- WCAG 2.1 AA compliance verification
- Screen reader compatibility
- Color blindness testing

### Phase 6: Documentation & Migration Guide

#### 6.1 API Documentation
- Update OpenAPI spec with new theme JSON structure
- Document contrast rules and validation
- Provide example theme JSON configurations

#### 6.2 Developer Guide
- How to use component border radius variants
- How contrast enforcement works
- How to override defaults (if needed)

#### 6.3 Migration Guide
- Steps to migrate existing brands
- Default values for existing brands
- Breaking changes (if any)

## Implementation Steps

### Step 1: Backend Foundation (Week 1)
- [ ] Add contrast utility functions
- [ ] Update brand model with validation methods
- [ ] Update create/update use cases with validation
- [ ] Write unit tests

### Step 2: Frontend Theme System (Week 1-2)
- [ ] Update BrandThemeProvider to parse new structure
- [ ] Create contrast utility functions
- [ ] Apply CSS variables for component border radius
- [ ] Implement contrast enforcement

### Step 3: Component Migration (Week 2-3)
- [ ] Migrate Button component
- [ ] Migrate Card component
- [ ] Fix StatusChip contrast issue
- [ ] Migrate Badge, Input, Panel, Tile components
- [ ] Update Tailwind config

### Step 4: Testing & Validation (Week 3)
- [ ] Write component tests
- [ ] Visual regression testing
- [ ] Accessibility audit
- [ ] Cross-browser testing

### Step 5: Documentation (Week 4)
- [ ] Update API documentation
- [ ] Create developer guide
- [ ] Write migration guide
- [ ] Update branding settings UI (if needed)

## Example: Fixed StatusChip

**Before** (Current - has contrast issue):
```tsx
styles: "bg-cyan-100/80 text-cyan-700 ..."  // Cyan text on cyan background
```

**After** (Fixed):
```tsx
const bgColor = "bg-cyan-100/80"
const textColor = ensureContrast(bgColor, "text-cyan-700", { minRatio: 4.5 })
styles: `${bgColor} ${textColor} ...`  // Auto-adjusted for contrast
```

## Example: Button with Theme Border Radius

**Before**:
```tsx
className="rounded-full"  // Hardcoded
```

**After**:
```tsx
className="rounded-[var(--border-radius-button-default)]"  // Theme-controlled
// Or with variant support:
className={cn(
  variant === 'rounded' && "rounded-[var(--border-radius-button-rounded)]",
  variant === 'square' && "rounded-[var(--border-radius-button-square)]",
  "rounded-[var(--border-radius-button-default)]"
)}
```

## Success Criteria

1. ✅ All components use theme-controlled border radius
2. ✅ No contrast violations (all text readable)
3. ✅ WCAG 2.1 AA compliance (4.5:1 contrast ratio)
4. ✅ Backward compatible (defaults for existing brands)
5. ✅ All components support multiple corner styles (round, rounded, square)
6. ✅ Automatic contrast adjustment when needed
7. ✅ Comprehensive test coverage

## Risks & Mitigation

### Risk 1: Breaking Changes
- **Mitigation**: Provide sensible defaults, backward compatible migration

### Risk 2: Performance Impact
- **Mitigation**: Cache contrast calculations, use CSS variables (no runtime cost)

### Risk 3: Visual Inconsistencies
- **Mitigation**: Comprehensive testing, visual regression tests

### Risk 4: Complex Theme JSON
- **Mitigation**: Provide UI for theme configuration, clear documentation

## Future Enhancements

1. **Theme Presets**: Pre-defined theme configurations (Modern, Classic, Minimal)
2. **Visual Theme Builder**: UI for configuring all theme options
3. **Real-time Preview**: See theme changes before saving
4. **Export/Import**: Share theme configurations between brands
5. **Advanced Contrast Options**: Custom contrast ratios per component type
