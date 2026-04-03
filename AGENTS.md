# Agent Notes

## Error handling style

- Do not use inline error declarations in conditionals:
  - Avoid: `if err := doSomething(); err != nil { ... }`
- Use explicit assignment before the check:
  - Prefer:
    - `err := doSomething()`
    - `if err != nil { ... }`
