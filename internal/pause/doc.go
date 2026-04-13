// Package pause provides a Controller that allows pipeline components to be
// temporarily halted and resumed without stopping the process.
//
// Typical usage:
//
//	ctrl := pause.New()
//
//	// In the processing loop:
//	if err := ctrl.Wait(ctx); err != nil {
//	    return err
//	}
//
//	// From an operator or signal handler:
//	ctrl.Pause()
//	ctrl.Resume()
package pause
