# Implementing Pipelines

Notes on how to add `|` support to this shell.

## 1. Split on `|` inside the tokenizer, not with `strings.Split`

`strings.Split(input, "|")` breaks on `echo "a|b"`. `parseArgs` already tracks
quote/escape state, so add one case to the unquoted branch:

```go
case r == '|':
    if hasToken { args = append(args, current.String()); current.Reset(); hasToken = false }
    cmds = append(cmds, args)   // flush the segment
    args = nil
```

Return `[][]string`. Everything quoted falls through to `default` as it already does.

**Worth considering while you're in there:** instead of `[][]string`, emit typed
tokens — `struct { value string; quoted bool; op bool }` — and split the token
stream on operators. That fixes a latent bug in the current redirect code:
`echo ">" file` currently redirects because `handle()` string-compares
`c.args[len-2]` without knowing whether the `>` was quoted. It also makes
`>` / `2>>` detection a scan over tokens rather than the fragile
"look at second-to-last arg" check.

## 2. A `Pipeline` type wrapping `[]*Command`

`NewPipeline(input) *Pipeline`, single command = pipeline of length 1 so the
existing path is unchanged. Move these up to pipeline level, since they're
properties of the whole line, not one command:

- trailing `&` (background)
- the `listJobs(c.stdout, true)` call at the end of `handle()`

Redirects stay per-command (each stage parses its own; an explicit `>` on a
stage overrides the pipe for that stage).

## 3. Use `os.Pipe()`, not `io.Pipe` or `StdoutPipe`

This matters a lot for `tail -f | head -n 5`. If you set `cmd.Stdout` to any
writer that isn't an `*os.File`, `os/exec` silently creates its own pipe plus a
goroutine copying bytes, and `Wait()` blocks on that goroutine. With `tail -f`
it never returns, and SIGPIPE never reaches `tail`. `os.Pipe()` gives you real
fds that get handed to the child directly:

```go
pr, pw, _ := os.Pipe()
cmds[0].Stdout = pw
cmds[1].Stdin  = pr
```

## 4. Start everything, then wait — and close the parent's fds

```go
for _, c := range cmds { c.Start() }
// CRITICAL: parent still holds copies of every pipe end
for _, f := range allPipeEnds { f.Close() }
for _, c := range cmds { c.Wait() }
```

The two classic bugs here, both of which the tester will hit:

- **Forgetting the parent's `pw.Close()`** → `wc` never sees EOF and hangs
  forever, because the parent still has the write end open even after `cat` exits.
- **Running stages sequentially** (`cmd1.Run()` then `cmd2.Run()`) → deadlocks as
  soon as output exceeds the 64K pipe buffer, and `tail -f` never terminates at all.

## 5. Exit status and the broken-pipe case

Pipeline status = **last** command's status. `cat`'s `Wait()` in
`tail -f | head -n 5` will return `signal: broken pipe` / SIGPIPE once `head`
exits — that's expected, not an error to report. Still call `Wait()` on every
stage though, or you leak zombies (and `jobMap` reaping gets confused).

Go children get default SIGPIPE disposition (the kernel resets handlers across
`exec`), so `tail` dies on its own once `head` closes the read end. Nothing
special needed.

## 6. Builtins in a pipeline

Not required by this stage, but cheap to design for since the builtins already
write to `c.stdout io.Writer`. Run the builtin in a goroutine writing to the
pipe's write end, closing it when done. Two edge cases: `exit` and `cd` in a
pipeline run in a subshell in real bash, so they shouldn't kill or move the shell.

## Suggested order

1. Tokenizer change + `parseArgs` → `[][]string` (or tokens), single-command
   path still green.
2. `Pipeline` struct, `handle()` becomes a loop that just runs stage 0 when `len == 1`.
3. Pipe wiring + start-all/close-all/wait-all.
4. Move `&` and job reaping to pipeline level.

Fix the redirect parsing before the wiring — right now it's coupled to arg
positions, and once a command is a pipeline stage, `cmd | grep foo > out.txt`
gets ambiguous fast.
