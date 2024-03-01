
[![tweet-1757420590103732371](https://github.com/Aslemammad/deconf/assets/37929992/8bb77b1d-4ac6-4f06-a75c-b3008b500aab)](https://twitter.com/_pi0_/status/1757420590103732371)
# Deconf (RFC)
This is a trial for unifying the tooling configuration in the javascript ecosystem without breaking the current ecosystem or requiring a heavy change from the library authors. 

There are also other attempts in this direction, check [config-dir](https://github.com/pi0/config-dir).

## Syntax 
The overall goal is to have all of our configurations in one file. As discussed in Twitter, JSON and Markdown can work.

Markdown looks ideal since it would support code blocks (e.g. js) and the only work that needs to be done is parsing the markdown and then linking those language markdown blocks to the tools we use daily. 

`config.md`:

```md
# configuration

## `vite.config.ts`

\`\`\`ts
import { defineConfig } from 'vite'

export default defineConfig({
  esbuild: {
    jsxFactory: 'h',
    jsxFragment: 'Fragment',
  },
})
\`\`\`

Ignore the triple backslashes (\) as they are being used here to avoid github's markdown bug. In the real-world use case, triple backticks (`) are enough.

## `.eslintrc.json`

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



The frontmatter section (optional) maps each second level heading to the corresponding file path on the disk. In this example, it maps the vite configuration to the `vite.config.ts`. 

Second level headings also accept files directly without the need for frontmatter by just wrapping their value with __\`\`__ similar to the __\`tsconfig.json\`__ configuration.

Other levels of headings and markdown features are basically ignored so the user can use them for documenting or commenting on configurations.

The languages in code blocks (**ts** in __\`\`\`ts__) are also ignored by the parser but they can be helpful for the user's editor to enable syntax highlighting or formatting.

## Implementation
### Symlinks
Deconf has to parse the markdown and for each configuration, it creates a corresponding file in `node_modules/.deconf` and then a symlink in the project's root directory. 

The reason it should be included in the directory is that the purpose of Deconf is to not break the ecosystem conventions.
```
/vite.config.ts => /node_modules/.deconf/vite.config.ts
```

### Workspace

The symbolic link files should ideally be ignored by Git and the editor, possibly through an automatic command in Deconf.
### Changes

Deconf needs to monitor changes in the `config.md` file and replicate the necessary changes on the user's disk.

As there is no npm hook for running before any npm command, a potential solution might involve initiating a process during the system's startup to monitor the `config.md` files that were initialized by a command such as `deconf init`.

For other users that haven't initialized the `config.md` (with `deconf init`) themselves might need a `postinstall` script in the project so the startup process identifies the `config.md` files in the project.
## Lsp

The concern is the language server support for the code blocks in Markdown, enabling users to access features like autocompletion for JavaScript or TypeScript code blocks in the `config.md` file.
So more information is needed for this section. 


## TODO 
- [x] parse config md
- [x] emit files in .deconf
- [x] symlinking
- [ ] git ignore modify command
- [ ] vscode ignore (https://stackoverflow.com/a/36193178)
- [ ] daemon
  - [ ] add files to watchlist in daemon
  - [ ] run daemon in the start of system (.zshrc?)

  
## Contributing 

Feel free to raise an issue about the design or add any implementation detail to any of the sections above. 
