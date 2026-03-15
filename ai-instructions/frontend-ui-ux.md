# Frontend UI/UX Design — Agent Instructions

> **Scope:** Frontend
> These UI/UX standards **must be strictly followed** whenever you create, modify, or review any frontend design or interface code, regardless of the nature of the work (new features, bug fixes, refactoring, or any other changes). If you find any contradiction between the rules in this file and those in any other instruction file (including `AGENTS.md`), **report it immediately and start a discussion** before making any changes.

---

## 1. Consistency & Simplicity

- Use the same colors, fonts, spacing scale, and icon set throughout the entire application. Never introduce one-off values; rely on the Tailwind CSS theme tokens already configured in the project.
- Keep all user-facing text short and direct: labels, button copy, empty-state messages, and error strings must all be concise and meaningful.
- Every page must contain only the interface elements required for its primary task. Remove any decorative or secondary element that does not directly help the user complete that task.
- Prefer familiar, conventional UI patterns (e.g., a top navigation bar, a standard form layout) over novel ones. Surprise the user as little as possible.

---

## 2. Accessibility (a11y)

- Use semantic HTML elements (`<nav>`, `<main>`, `<section>`, `<article>`, `<header>`, `<footer>`, `<button>`, `<a>`, etc.) wherever appropriate. Never use a `<div>` or `<span>` as an interactive element.
- Every interactive element must be reachable and operable via keyboard alone (Tab, Shift+Tab, Enter, Space, Escape, arrow keys where applicable).
- Provide descriptive `aria-label` or `aria-labelledby` attributes on controls whose visible label is insufficient or absent (e.g., icon-only buttons).
- Maintain a minimum contrast ratio of **4.5 : 1** for normal text and **3 : 1** for large text, following WCAG 2.1 AA guidelines.
- All images and icons that convey meaning must have an `alt` attribute (or `aria-hidden="true"` when purely decorative).
- Focus indicators must remain visible; never apply `outline: none` without providing an equivalent custom focus style.

---

## 3. Responsive & Mobile-Friendly Design

- All layouts must adapt gracefully from small mobile screens (≥ 320 px) up to wide desktop viewports. Use Tailwind's responsive prefixes (`sm:`, `md:`, `lg:`) to achieve this.
- Touch targets must be at least 44 × 44 px (following Apple HIG / Material Design guidance) to remain usable on touch devices.
- Navigation menus and sidebars must collapse or stack appropriately on small screens. Prefer a vertically scrollable single-column layout for narrow viewports.
- Never rely on hover-only interactions to reveal critical UI. Ensure the same information or action is accessible without hovering.

---

## 4. Navigation & Interaction

- Navigation must be shallow: users should reach any primary function in at most two interactions from any page.
- The current location in the app must always be visually indicated (e.g., active link highlight in the navigation bar).
- Provide immediate, visible feedback for every user action:
  - **Loading states:** Show a spinner, skeleton, or disabled state while an async operation is in progress.
  - **Success states:** Briefly display a confirmation (e.g., toast notification) when an action completes successfully.
  - **Error states:** Clearly surface error messages adjacent to the failing element, with guidance on how to resolve the issue.
- Destructive actions (delete, revoke) must require explicit confirmation before proceeding (e.g., a confirmation dialog).
- Links and buttons must have clear, action-oriented labels (`Create short link`, `Delete`, `Sign out`) — never generic ones (`Click here`, `Submit`).

---

## 5. Visual Hierarchy

- Use whitespace generously to group related elements and separate unrelated ones. Avoid dense, crowded layouts.
- Typography must establish a clear hierarchy: one prominent heading per page section, supporting body text at a readable size (minimum 16 px equivalent for body copy), and de-emphasised secondary text for metadata or captions.
- Use color and weight (bold) sparingly to draw attention only to the single most important call-to-action on a page. Do not make multiple elements compete visually.
- Avoid using more than two distinct font weights and one accent color for interactive elements on any given page.

---

## 6. Code Complexity Trade-offs

When implementing any of the above requirements would result in a **significant increase in template, TypeScript, or CSS complexity**, the agent must:

1. **Stop and communicate** the trade-off clearly to the user before writing the code.
2. **Propose up to three implementation options**, ranked from simplest to most complete, with a brief summary of the code complexity cost and the UX benefit of each.
3. **Wait for the user's decision** before proceeding. Do not choose an option unilaterally.

A "significant" trade-off means any of the following:

- Adding more than ~20 lines of template or logic purely for a UX enhancement.
- Requiring a new custom component or utility solely for an accessibility or animation feature.
- Introducing conditional rendering branches that materially affect readability of an existing component.

---

## 7. What Agents Must Not Do

- Do not introduce CSS animations or transitions that cannot be disabled via the `prefers-reduced-motion` media query.
- Do not hard-code color values directly in templates or styles; use Tailwind theme tokens or CSS variables already defined in the project.
- Do not implement a UX pattern that requires hover as the **only** means of discovery.
- Do not ship a page or component with interactive elements that are unreachable by keyboard.
- Do not add decorative UI elements, illustrations, or animations without explicit user approval.
