package gobo

type Options struct {
	ReplaceSlice bool
	AddNewSlice  bool
}

type Option func(*Options)

// If ReplaceSlice is true, it will replace the original slice with the new one.
// If it's false, it will conserve the differences of the new slice and the original one (default).
func UseReplaceSlice() Option {
	return func(opts *Options) {
		opts.ReplaceSlice = true
	}
}

// If FullSliceAdd is true, it will block ReplaceSlice property and conserve both original and new data
func UseAddNewSlice() Option {
	return func(opts *Options) {
		opts.AddNewSlice = true
	}
}
