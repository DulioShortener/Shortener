import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const stylesheet = await readFile(new URL('../src/styles/custom.css', import.meta.url), 'utf8')

test('the interactive reference escapes the prose width constraint', () => {
  assert.match(
    stylesheet,
    /\.sl-markdown-content\s*{\s*max-width:\s*72ch;\s*}/,
    'prose pages should retain a readable line length',
  )
  assert.match(
    stylesheet,
    /main:has\(\.scalar-api-reference\) \.sl-container,\s*main:has\(\.scalar-api-reference\) \.sl-markdown-content\s*{\s*max-width:\s*none;\s*}/,
    'the embedded API reference should use the available viewport width',
  )
})
