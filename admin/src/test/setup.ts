// Vitest setup for admin unit tests (jsdom environment).
// Loads the Testing Library DOM matchers (toEqualText, toBeInTheDocument, ...)
// so component tests added later can use them out of the box.
import "@testing-library/jest-dom/vitest";
