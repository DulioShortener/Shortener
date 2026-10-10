---
name: Dulio API Documentation
description: A compact protocol handbook for understanding and exercising the Dulio Shortener API.
colors:
  accent-subtle: "#25264e"
  accent-action: "#7c83ff"
  accent-emphasis: "#d9dcff"
  canvas: "#111421"
  panel: "#202438"
  rule: "#353a50"
  text-strong: "#f8f9ff"
  text-body: "#c5c9dc"
  text-muted: "#969bb3"
  selection: "#5865f2"
  focus: "#8d94ff"
  light-canvas: "#ffffff"
  light-panel: "#f5f6fa"
  light-rule: "#c7cad7"
  light-text-strong: "#171a2b"
  light-text-body: "#454a63"
  light-accent-subtle: "#e6e8ff"
  light-accent-action: "#4f56c9"
  light-accent-emphasis: "#20256f"
typography:
  display:
    fontFamily: "Aptos, Segoe UI Variable, Segoe UI, system-ui, sans-serif"
    fontSize: "2.625rem"
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: "-0.025em"
  headline:
    fontFamily: "Aptos, Segoe UI Variable, Segoe UI, system-ui, sans-serif"
    fontSize: "2.1875rem"
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: "-0.025em"
  title:
    fontFamily: "Aptos, Segoe UI Variable, Segoe UI, system-ui, sans-serif"
    fontSize: "1.8125rem"
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: "-0.025em"
  body:
    fontFamily: "Aptos, Segoe UI Variable, Segoe UI, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.75
    letterSpacing: "normal"
  label:
    fontFamily: "Aptos, Segoe UI Variable, Segoe UI, system-ui, sans-serif"
    fontSize: "0.82rem"
    fontWeight: 700
    lineHeight: 1.2
    letterSpacing: "0.04em"
  code:
    fontFamily: "Cascadia Code, SFMono-Regular, Consolas, monospace"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.65
    letterSpacing: "normal"
rounded:
  xs: "0.25rem"
  inline: "0.35rem"
  panel: "0.45rem"
  control: "0.5rem"
spacing:
  2xs: "0.25rem"
  xs: "0.5rem"
  sm: "0.75rem"
  md: "1rem"
  lg: "1.5rem"
components:
  search-trigger:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.text-body}"
    typography: "{typography.body}"
    rounded: "{rounded.control}"
    padding: "0.625rem 0.5rem 0.625rem 0.75rem"
    height: "2.5rem"
  sidebar-active:
    backgroundColor: "{colors.accent-emphasis}"
    textColor: "{colors.canvas}"
    typography: "{typography.body}"
    rounded: "{rounded.xs}"
    padding: "0.25rem 0.5rem"
  inline-code:
    backgroundColor: "{colors.rule}"
    textColor: "{colors.text-strong}"
    typography: "{typography.code}"
    rounded: "{rounded.inline}"
    padding: "0.12em 0.35em"
  code-panel:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.text-strong}"
    typography: "{typography.code}"
    rounded: "{rounded.panel}"
    padding: "1rem"
  tip-aside:
    backgroundColor: "{colors.accent-subtle}"
    textColor: "{colors.text-strong}"
    typography: "{typography.body}"
    padding: "1rem"
  method-badge:
    backgroundColor: "{colors.rule}"
    textColor: "{colors.text-body}"
    typography: "{typography.code}"
    rounded: "{rounded.xs}"
    padding: "0.125rem 0.375rem"
---

# Design System: Dulio API Documentation

## Overview

**Creative North Star: "The Protocol Desk"**

The Dulio documentation is a compact protocol handbook: an ink-blue working surface where the public contract is easier to find than any pitch. Its familiar developer-reference architecture uses persistent navigation, a quiet reading column, a compact page outline, and a direct path into the interactive reference.

The atmosphere is technical, calm, and dense without becoming cramped. Electric blurple identifies action and orientation; crisp rules, code surfaces, method badges, and tabular numerals do most of the structural work. The system deliberately rejects marketing gloss, ornamental gradients, oversized hero composition, and bespoke MDX machinery that would obscure a small API.

**Key Characteristics:**

- Ink-blue, dark-first reading surfaces with a complete restrained light theme.
- Electric blurple reserved for links, active states, selection, and focus.
- Compact three-region documentation architecture on wide screens.
- System sans typography paired with a native technical monospace stack.
- Flat, ruled containers and code-led examples instead of decorative cards.

## Colors

The palette is a cool ink scale interrupted by a narrow electric-blurple signal; the light theme inverts the same hierarchy without becoming a separate visual identity.

### Primary

- **Electric Blurple:** The action color for links, active navigation, interactive emphasis, and the selected-text family.
- **Signal Mist:** The high-contrast accent used for readable active states on ink surfaces.
- **Deep Blurple:** A low-luminance accent field used behind selected or advisory content.

### Neutral

- **Midnight Ink:** The default dark canvas and the anchor for the site's focused reading atmosphere.
- **Protocol Panel:** The raised-by-tone surface for navigation, code examples, and dense utility regions.
- **Crisp Rule:** The border and divider color that separates structure without adding decorative chrome.
- **Paper White:** The strongest dark-theme text, reserved for headings and high-priority labels.
- **Cool Silver:** The default dark-theme body text.
- **Muted Slate:** Secondary copy, table-of-contents links, and quiet metadata.
- **Paper Canvas / Paper Panel:** The light-theme page and utility surfaces.
- **Paper Ink / Paper Body / Paper Rule:** The corresponding light-theme hierarchy for text and dividers.

### Named Rules

**The Signal, Not Wallpaper Rule.** Electric blurple marks something actionable, selected, or focused; it is not a decorative background wash.

**The Two Complete Themes Rule.** Dark is the visual lead, but every semantic role must resolve deliberately in light mode rather than relying on automatic inversion.

## Typography

**Display Font:** Aptos (with Segoe UI Variable, Segoe UI, system-ui, and sans-serif fallbacks)

**Body Font:** Aptos (with Segoe UI Variable, Segoe UI, system-ui, and sans-serif fallbacks)

**Label/Mono Font:** Cascadia Code (with SFMono-Regular, Consolas, and monospace fallbacks)

**Character:** A practical Windows-first system stack keeps the handbook fast and familiar. Tight heading tracking gives the sans typography authority, while the monospace face makes identifiers, methods, paths, payloads, and examples immediately scannable.

### Hierarchy

- **Display** (600, 2.625rem on wide screens, 1.2 line-height): Page titles only; it contracts to 2.1875rem below the medium breakpoint.
- **Headline** (600, 2.1875rem on wide screens, 1.2 line-height): Major document sections; it contracts to 1.8125rem on narrow screens.
- **Title** (600, 1.8125rem on wide screens, 1.2 line-height): Subsections and strong local headings.
- **Body** (400, 1rem, 1.75 line-height): Explanatory copy in a reading measure capped at 72ch.
- **Label** (700, 0.82rem, 0.04em tracking, uppercase): Request/response captions and other compact protocol labels.
- **Code** (400, 0.875rem, 1.65 line-height): General code; the compact paired request example reduces to 0.72rem to keep real requests legible side by side.

### Named Rules

**The Contract Leads Rule.** Use monospace only for literal protocol material; headings and explanatory prose remain in the system sans stack.

**The Calm Density Rule.** Preserve the generous body line-height and 72ch measure even when the surrounding navigation and labels are compact.

## Layout

The wide-screen shell is a three-region reference layout: a fixed 18.75rem navigation rail, a reading column, and an on-page outline separated by one-pixel rules. The header is 3.5rem tall on narrow screens and 4rem from 50rem upward. Content padding begins at 1rem and grows to 1.5rem at 72rem.

At 72rem, the right-hand outline becomes persistent; below that width it collapses to a compact mobile table of contents. At 50rem, the left navigation changes from a full-width popover into the fixed rail. The signature request/response example uses two equal fluid columns with a 1rem gutter and collapses to one column at 50rem. Tables remain horizontally scrollable rather than compressing protocol values into unreadable wrapping.

Spacing follows a compact quarter-rem rhythm, with 1rem as the dominant content and container interval. Document blocks use a consistent 1rem vertical gap, while a heading that follows body content receives additional separation.

**The Reference Before Promotion Rule.** The first viewport establishes navigation, the production base URL, and an executable request; it does not spend space on a marketing hero.

## Elevation & Depth

The system is flat by default. Depth comes from tonal layering and crisp one-pixel rules: canvas, navigation panel, code panel, and active state each occupy a distinct value. Inherited Starlight shadows are reserved for floating or navigational utilities such as the mobile menu, search dialog, and previous/next pagination; reading content and protocol examples remain unshadowed.

### Shadow Vocabulary

- **Utility low** (`0 1px 1px #0000001f, 0 2px 1px #0000003d`): Small floating controls on the dark theme.
- **Utility medium** (`0 8px 4px #00000014, 0 5px 2px #00000014, 0 3px 2px #0000001f, 0 1px 1px #00000026`): Pagination and compact navigation utilities.
- **Modal high** (`0 25px 7px #00000008, 0 16px 6px #0000001a, 0 9px 5px #16181d54, 0 4px 4px #000000bf, 0 4px 2px #00000040`): Search dialog only.

### Named Rules

**The Flat Contract Rule.** Never use a shadow to make ordinary documentation content feel important; use hierarchy, rules, and tone.

## Shapes

Corners are tight and utilitarian. Inline code uses a gently clipped 0.35rem radius, protocol panels use 0.45rem, and interactive controls top out at 0.5rem. Circular geometry is reserved for the mobile menu control. Borders are one-pixel structural rules, not decorative frames, and code or table content may scroll inside its container rather than distorting the page.

**The Crisp Edge Rule.** Avoid large soft cards and pill-heavy interfaces; compact radii should support scanning, not advertise friendliness.

## Components

### Buttons

- **Shape:** Compact controls use a 0.5rem radius; the mobile menu is the circular exception.
- **Primary:** Interactive-reference actions use electric blurple or the embedded reference's high-contrast action treatment; reserve filled actions for operations readers can execute.
- **Hover / Focus:** Hover strengthens border or text contrast. Every keyboard target receives the global 3px focus outline with a 3px offset.
- **Secondary / Ghost:** Header utilities remain transparent or canvas-toned so they do not compete with the reference content.

### Chips

- **Style:** Method badges and compact protocol markers use monospace text on a ruled ink surface with a 0.25rem radius.
- **State:** Color communicates operation or status only when it remains readable in both themes; labels never rely on color alone.

### Cards / Containers

- **Corner Style:** Flat protocol panels use the 0.45rem panel radius; navigation/pagination utilities may use 0.5rem.
- **Background:** Use the protocol panel tone above the midnight canvas, or its paper-theme equivalent.
- **Shadow Strategy:** Content containers have no shadow; floating utilities use the elevation vocabulary.
- **Border:** One crisp rule defines code, tables, search, pagination, and section boundaries.
- **Internal Padding:** 1rem is the default; dense inline tokens use fractional-em padding.

### Inputs / Fields

- **Style:** Search fields use a canvas background, one-pixel rule, and compact 0.5rem corners.
- **Focus:** Shift the rule to electric blurple and retain the global visible outline where the native control receives focus.
- **Error / Disabled:** Preserve readable text and explicit state language; never encode a failure with hue alone.

### Navigation

The fixed rail uses compact nested lists, subdued default text, and a high-contrast blurple-tinted active row. The top header keeps the product title, search, repository link, and theme control aligned in one quiet utility band. On narrow screens the rail becomes a keyboard-trapped popover and the page outline becomes a compact disclosure.

### Quick Request Pair

This is the signature component: request and response appear as equal code panels with uppercase captions, real protocol values, synchronized height, and independent overflow. It becomes a vertical sequence below 50rem, keeping the request before the response in source order.

### Advisory Aside

Tips use a narrow semantic edge, tinted ink field, icon, short title, and body copy. They direct readers toward the interactive reference without turning into a promotional banner.

## Do's and Don'ts

### Do:

- **Do** surface base URLs, authentication, limits, and concrete requests before secondary explanation.
- **Do** preserve the 72ch reading measure, 1.75 body line-height, and explicit wide/mobile navigation transitions.
- **Do** use electric blurple for action, focus, selection, and orientation.
- **Do** keep examples copyable, horizontally scrollable where necessary, and numerically aligned.
- **Do** test every component in both the ink and paper themes.

### Don't:

- **Don't** add a marketing hero, oversized slogan, testimonial, or unsupported product claim.
- **Don't** replace crisp rules and tonal layers with glossy cards, ambient gradients, or ubiquitous shadows.
- **Don't** introduce a web-font dependency when the system stacks already express the intended technical calm.
- **Don't** invent MDX-only abstractions for patterns that plain Markdown, semantic HTML, or the existing Starlight primitives already handle.
- **Don't** collapse method, path, status, or error meaning into color alone.
