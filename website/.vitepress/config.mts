import { defineConfig } from 'vitepress'

// GitHub Pages serves the site from /<repo-name>/ — override with DOCS_BASE
// when building for a custom domain or local preview.
const base = process.env.DOCS_BASE || '/portfolio/'
const repoUrl = 'https://github.com/shafi-/portfolio'

export default defineConfig({
  lang: 'en-US',
  title: 'Portfolio Docs',
  description:
    'Portfolio is a local-first project inventory and knowledge platform for developers and AI coding agents. Discover every repository, extract its metadata, index its docs — 100% local and offline.',

  base,
  srcDir: '.',
  outDir: '.vitepress/dist',
  cacheDir: '.vitepress/cache',

  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: `${base}favicon.svg` }],
    ['link', { rel: 'preconnect', href: 'https://fonts.googleapis.com' }],
    ['link', { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' }],
    [
      'link',
      {
        rel: 'stylesheet',
        href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap'
      }
    ]
  ],

  markdown: {
    theme: { light: 'github-light', dark: 'github-dark' },
    lineNumbers: false
  },

  themeConfig: {
    logo: '/logo.svg',
    siteTitle: 'Portfolio',

    nav: [
      { text: 'Docs', link: '/docs/getting-started/introduction', activeMatch: '^/docs/' },
      { text: 'CLI', link: '/docs/cli/', activeMatch: '^/docs/cli/' },
      { text: 'HTTP API', link: '/docs/api/', activeMatch: '^/docs/api/' },
      { text: 'MCP', link: '/docs/mcp/', activeMatch: '^/docs/mcp/' },
      {
        text: 'v0.3.4',
        items: [
          { text: 'Release notes', link: '/docs/release-notes' },
          { text: 'All releases', link: `${repoUrl}/releases` }
        ]
      }
    ],

    sidebar: {
      '/docs/': [
        {
          text: 'Prologue',
          items: [
            { text: 'Release Notes', link: '/docs/release-notes' },
            { text: 'Contribution Guide', link: '/docs/contribution-guide' }
          ]
        },
        {
          text: 'Getting Started',
          items: [
            { text: 'Introduction', link: '/docs/getting-started/introduction' },
            { text: 'Installation', link: '/docs/getting-started/installation' },
            { text: 'Quick Start', link: '/docs/getting-started/quick-start' },
            { text: 'Configuration', link: '/docs/getting-started/configuration' }
          ]
        },
        {
          text: 'Core Concepts',
          items: [
            { text: 'Discovery', link: '/docs/concepts/discovery' },
            { text: 'Workspaces', link: '/docs/concepts/workspaces' },
            { text: 'Metadata Extraction', link: '/docs/concepts/metadata' },
            { text: 'The Knowledge Store', link: '/docs/concepts/knowledge-store' },
            { text: 'Architecture', link: '/docs/concepts/architecture' }
          ]
        },
        {
          text: 'CLI Reference',
          items: [
            { text: 'Available Commands', link: '/docs/cli/' },
            { text: 'Projects', link: '/docs/cli/projects' },
            { text: 'Workspaces', link: '/docs/cli/workspaces' },
            { text: 'Scanning & Status', link: '/docs/cli/scanning' },
            { text: 'Configuration Commands', link: '/docs/cli/configuration' },
            { text: 'AI Integrations', link: '/docs/cli/ai-integrations' },
            { text: 'Servers & Tools', link: '/docs/cli/servers' }
          ]
        },
        {
          text: 'HTTP API',
          items: [
            { text: 'API Overview', link: '/docs/api/' },
            { text: 'Endpoints', link: '/docs/api/endpoints' }
          ]
        },
        {
          text: 'MCP Integration',
          items: [
            { text: 'MCP Server', link: '/docs/mcp/' },
            { text: 'Connecting AI Agents', link: '/docs/mcp/agents' },
            { text: 'Tools Reference', link: '/docs/mcp/tools' }
          ]
        },
        {
          text: 'Dashboard',
          items: [{ text: 'Dashboard', link: '/docs/dashboard' }]
        },
        {
          text: 'Digging Deeper',
          items: [
            { text: 'Database & Security', link: '/docs/digging-deeper/security' },
            { text: 'Troubleshooting', link: '/docs/digging-deeper/troubleshooting' },
            { text: 'Release Process', link: '/docs/digging-deeper/release-process' }
          ]
        }
      ]
    },

    search: {
      provider: 'local',
      options: {
        translations: {
          button: { buttonText: 'Search docs', buttonAriaLabel: 'Search documentation' }
        }
      }
    },

    socialLinks: [{ icon: 'github', link: repoUrl }],

    outline: { level: [2, 3], label: 'On this page' },

    docFooter: { prev: 'Previous', next: 'Next' },

    footer: {
      message: 'Released under the GNU AGPL-3.0 license.',
      copyright: 'Copyright © 2026 Portfolio Contributors'
    }
  }
})
