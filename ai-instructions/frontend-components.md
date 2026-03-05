# Frontend Components — Agent Instructions

> **Scope:** Frontend
> These coding standards and conventions **must be strictly followed** whenever you work on any frontend UI code, regardless of the nature of the work (new features, bug fixes, refactoring, or any other changes). If you find any contradiction between the rules in this file and those in any other instruction file (including `AGENTS.md`), **report it immediately and start a discussion** before making any changes.

---

## 1. Core Rule — shadcn/vue Is the Only UI Component Library

**Every UI element in this application must use a shadcn/vue component.** There are no exceptions unless the user explicitly instructs you to create a custom component for a specific, named use case.

- **Never create custom UI components on your own initiative.** If a task appears to require a UI element not yet available as a shadcn/vue component, first check whether it can be composed from existing shadcn/vue components before concluding a custom component is necessary.
- **If and only if** the user explicitly instructs you to create a custom component, place it in `src/custom-components/` — **never** in `src/components/`, which is reserved exclusively for shadcn/vue components installed via the CLI.
- Do not wrap, subclass, or otherwise extend shadcn/vue components into new components without explicit user approval.

---

## 2. Installing shadcn/vue Components

- Add components using the shadcn/vue CLI (`npx shadcn-vue@latest add <component>`). Do not copy component source manually.
- Do not modify the source of installed shadcn/vue components in `src/components/`. They are used **as-is**. If a styling or behaviour change seems necessary, check whether it can be achieved through props, slots, or Tailwind utility classes in the consuming template first.
- If you encounter a situation where the "as-is" constraint causes significant code duplication or complexity, **report it immediately** and start a discussion before modifying any shadcn/vue component source.

---

## 3. Composing UI from shadcn/vue Primitives

- Prefer composing multiple shadcn/vue primitives (e.g., `Dialog`, `DropdownMenu`, `Card`, `Button`) directly in view or parent components rather than wrapping them in new single-purpose wrapper components.
- Page layout and spacing should be achieved with Tailwind CSS utility classes applied directly in `<template>`, not with additional wrapper components.
- Use shadcn/vue's slot-based API to inject custom content into components (e.g., `DialogContent`, `CardHeader`) rather than duplicating template structure in separate files.

---

## 4. Directory Conventions

| Path                     | Purpose                                                          |
| ------------------------ | ---------------------------------------------------------------- |
| `src/components/`        | shadcn/vue components installed via CLI — **no custom code**     |
| `src/custom-components/` | Custom components created **only** on explicit user instruction  |
| `src/views/`             | Page-level components (one per route); may use both of the above |

---

## 5. Component Authoring Standards

- All Vue components (`src/views/` and `src/custom-components/`) must use `<script setup lang="ts">`. The Options API is not permitted.

---

## 6. Styling

- Use Tailwind CSS utility classes exclusively. Do not write custom CSS anywhere except `src/style.css` for global base styles.
- If you think that using styling inside a Vue component for some particular case would allow you to make code considerably more concise and/or readable or dramatically reduce code duplication, you must ask the user for permission to do so. You may do it IF AND ONLY IF the user explicitly approves it or asks you to do it. If you find any contradiction about this requirement, report it immediately and start a discussion.

---

## 7. Testing

- shadcn/vue components installed into `src/components/` are excluded from test coverage requirements, as they are used as-is without customization. They may still be covered by tests when the consuming component's logic requires it (e.g., conditional rendering based on props). If you find any contradiction about it, report it immediately and start a discussion.

---

## 8. What Agents Must Not Do

- Do not create any new file under `src/components/` manually; only the shadcn/vue CLI may place files there.
- Do not build custom form inputs, buttons, modals, tables, tooltips, badges, or any other UI primitive from scratch when a shadcn/vue equivalent exists.
- Do not introduce any additional UI component library (e.g., Vuetify, PrimeVue, Element Plus, Headless UI) — shadcn/vue and Tailwind CSS are the only permitted UI tools.
- Do not use inline `style` attributes for static values; express all styling through Tailwind CSS utility classes.
