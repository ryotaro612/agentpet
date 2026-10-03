# agentpet

agentpet is a mcp server and a client to display spritesheets using a web view.Create a Makefile that runs 

## Zsh completion

Load completion for the current shell:

```bash
source <(petowner completion zsh)
```

To load it automatically, add that line to `~/.zshrc`. If `petowner` is not in
`PATH`, use its path both when loading and running it:

```zsh
source <(./dist/petowner completion zsh)
./dist/petowner -p 60012 pet <Tab>
```

Pet and animation names are retrieved from the running Agentpet server when
`-p` is specified. The completion function reuses the command path as typed,
so dynamic completion does not require `petowner` to be in `PATH`.

## Memo
- [ ] list_pets respnse 
- [ ] server fields
- [ ] remove animNameTools
