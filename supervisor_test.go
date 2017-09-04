package supervisor

// func Test01(t *testing.T) {
// 	var sum int64
// 	s, _ := Supervise(func() error {
// 		atomic.AddInt64(&sum, 1)
// 		return errorf("DUMMY")
// 	}, 3, time.Millisecond*50)
// 	s()
// 	assert.Equal(t, int64(3), sum)
// }
