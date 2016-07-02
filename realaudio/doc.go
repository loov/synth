// package realaudio provides easy rendering of realtime audio
//
//      err := realaudio.Run(
//      	// optional initialization func
//      	func(format realaudio.Format) error {
//      		// initialize your necessary buffers
//      		return nil
//      	},
//      	func(buffer []float32) error {
//      		// render your sound into data
//       		//   when done return realaudio.ErrDone
//       		return nil
//      	},
//      )
//

package realaudio
