
[![tweet-1757420590103732371](https://github.com/Aslemammad/deconf/assets/37929992/8bb77b1d-4ac6-4f06-a75c-b3008b500aab)](https://twitter.com/_pi0_/status/1757420590103732371)
# Deconf
This is a trial for unifying the toolling configuration in the javascript ecosytem without breaking the current ecosystem or requiring a heavy change from the library authors. 

There are also other attempts in this direction, check [config-dir](https://github.com/pi0/config-dir).

## Syntax 
The overall goal is to have all of our configurations in one file. As discussed in Twitter, json and md can work but md looks ideal since it'd support code blocks (e.g. js) and the only work that needs to be done is parsing the markdown and then linking those language markdown blocks to the tools we use daily. 

```md
---
vite: vite.config.ts
eslint: .eslintrc.json
---

# configuration

## vite

\`\`\`ts
import { defineConfig } from 'vite'

export default defineConfig({
  esbuild: {
    jsxFactory: 'h',
    jsxFragment: 'Fragment',
  },
})
\`\`\`

## eslint

\`\`\`json
{
    "rules": {
        "eqeqeq": "warn",
        "strict": "off"
    }
}
\`\`\`

## `tsconfig.json`

\`\`\`json
{
  "compilerOptions": {
    "module": "system",
    "noImplicitAny": true,
    "removeComments": true,
    "preserveConstEnums": true,
    "outFile": "../../built/local/tsc.js",
    "sourceMap": true
  },
  "include": ["src/**/*"],
  "exclude": ["**/*.spec.ts"]
}
\`\`\`
```

The frontmatter section (optional) maps each second level heading to the corresponding file on the disk. In this example, it maps the vite configuration to the `vite.config.ts`. 

Second level headings also accept files directly without the need for frontmatter by just wrapping their value with __\`\`__ similar to the __\`tsconfig.json\`__ configuration.

Other levels of headings and markdown features are basically ignored so the user can use them for documenting or commenting on configurations.

The languges in code blocks (**ts** in __\`\`\`ts__) are also ignored by the parser but they can be helpful for the user's editor to enable in-markdown syntax highlighting or formatting.

## Implementation

## Lsp