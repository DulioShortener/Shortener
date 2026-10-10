import { defineConfig } from 'astro/config'
import starlight from '@astrojs/starlight'
import { scalarStarlight } from '@scalar/starlight'

export default defineConfig({
  site: 'https://api-reference.3dreamstudio.com.br',
  integrations: [
    starlight({
      title: 'Dulio API',
      description: 'The complete API reference for Dulio Shortener.',
      favicon: '/favicon.svg',
      disable404Route: true,
      customCss: ['./src/styles/custom.css'],
      editLink: {
        baseUrl: 'https://github.com/DulioShortener/Shortener/edit/master/docs/',
      },
      lastUpdated: true,
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/DulioShortener/Shortener',
        },
      ],
      sidebar: [
        {
          label: 'Start here',
          items: [
            { label: 'Overview', slug: 'index' },
            { label: 'Getting started', slug: 'guides/getting-started' },
          ],
        },
        {
          label: 'Guides',
          items: [
            { label: 'Authentication', slug: 'guides/authentication' },
            { label: 'IDs and timestamps', slug: 'concepts/ids-and-timestamps' },
            { label: 'Links and URLs', slug: 'concepts/links-and-urls' },
          ],
        },
        {
          label: 'Protocol',
          items: [
            { label: 'Data models', slug: 'reference/data-models' },
            { label: 'Routing and limits', slug: 'reference/routing-and-limits' },
            { label: 'Errors', slug: 'reference/errors' },
          ],
        },
      ],
      plugins: [
        scalarStarlight({
          pathname: '/api-reference',
          label: 'Interactive API reference',
          title: 'Interactive API reference',
          configuration: {
            url: '/openapi.yaml',
            theme: 'purple',
            withDefaultFonts: false,
            hideTestRequestButton: false,
            showDeveloperTools: 'never',
            mcp: {
              disabled: true,
            },
            customCss: `
              .references-sidebar .flex.gap-1\\.5 > button:last-of-type,
              .agent-button-container,
              .scalar-mcp-layer-link {
                display: none !important;
              }
            `,
            authentication: {
              preferredSecurityScheme: 'bearerAuth',
            },
          },
        }),
      ],
    }),
  ],
})
