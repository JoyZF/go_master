# 一种类似装饰器的设计模式
```go

// runExitHooks runs any registered exit hook functions (funcs
// previously registered using runtime.addExitHook). Here 'exitCode'
// is the status code being passed to os.Exit, or zero if the program
// is terminating normally without calling os.Exit).
func runExitHooks(exitCode int) {
	if exitHooks.runningExitHooks {
		throw("internal error: exit hook invoked exit")
	}
	exitHooks.runningExitHooks = true

	runExitHook := func(f func()) (caughtPanic bool) {
		defer func() {
			if x := recover(); x != nil {
				caughtPanic = true
			}
		}()
		f()
		return
	}

	finishPageTrace()
	for i := range exitHooks.hooks {
		h := exitHooks.hooks[len(exitHooks.hooks)-i-1]
		if exitCode != 0 && !h.runOnNonZeroExit {
			continue
		}
		if caughtPanic := runExitHook(h.f); caughtPanic {
			throw("internal error: exit hook invoked panic")
		}
	}
	exitHooks.hooks = nil
	exitHooks.runningExitHooks = false
}

```