/**
 * Utility functions to parse API responses
 */

/**
 * Parse comma-separated string into array of strings
 * Handles empty strings and trims whitespace
 */
export function parseCommaSeparated(value: string | string[]): string[] {
  if (Array.isArray(value)) {
    return value
  }

  if (!value || value.trim() === '') {
    return []
  }

  return value.split(',').map(item => item.trim()).filter(item => item !== '')
}

/**
 * Parse language summary from API response
 */
export function parseLanguageSummary(languageSummary: string): string[] {
  return parseCommaSeparated(languageSummary)
}

/**
 * Parse framework summary from API response
 */
export function parseFrameworkSummary(frameworkSummary: string): string[] {
  return parseCommaSeparated(frameworkSummary)
}

/**
 * Parse dependency summary from API response
 */
export function parseDependencySummary(dependencySummary: string): string[] {
  return parseCommaSeparated(dependencySummary)
}
