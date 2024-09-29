---
Github Issue: https://github.com/nxtcoder17/Runfile/issues/8
---

### Expectations from this implementation ?

I would like to be able to do stuffs like:
- whether these environment variables have been defined 
- whether this filepath exists or not
- whether `this command` or script runs sucessfully

And, when answers to these questsions are `true`, then only run the target, otherwise throw the errors

### Problems with Taskfile.dev implementation

They have 2 ways to tackle this, with 

- `requires`: but, it works with vars only, ~no environment variables~

- `preconditions`: test conditions must be a valid linux `test` command, which assumes everyone knows how to read bash's `test` or `if` statements


### My Approach

1. Support for `test` commands, must be there, for advanced users

2. But, for simpler use cases, there should be alternate ways to do it, something that whole team just understands.
