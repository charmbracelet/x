package kitty

// Graphics image format.
const (
	// 32-bit RGBA format.
	RGBA = 32

	// 24-bit RGB format.
	RGB = 24

	// PNG format.
	PNG = 100
)

// Compression types.
const (
	Zlib = 'z'
)

// Transmission types.
const (
	// The data transmitted directly in the escape sequence.
	Direct = 'd'

	// The data transmitted in a regular file.
	File = 'f'

	// A temporary file is used and deleted after transmission.
	TempFile = 't'

	// A shared memory object.
	// For POSIX see https://pubs.opengroup.org/onlinepubs/9699919799/functions/shm_open.html
	// For Windows see https://docs.microsoft.com/en-us/windows/win32/memory/creating-named-shared-memory
	SharedMemory = 's'
)

// Action types.
const (
	// Transmit image data.
	Transmit = 't'
	// TransmitAndPut transmit image data and display (put) it.
	TransmitAndPut = 'T'
	// Query terminal for image info.
	Query = 'q'
	// Put (display) previously transmitted image.
	Put = 'p'
	// Delete image.
	Delete = 'd'
	// Frame transmits data for animation frames.
	Frame = 'f'
	// Animate controls animation.
	Animate = 'a'
	// Compose composes animation frames.
	Compose = 'c'
)

// Delete types.
const (
	// Delete all placements visible on screen
	DeleteAll = 'a'
	// Delete all images with the specified id, specified using the i key. If
	// you specify a p key for the placement id as well, then only the
	// placement with the specified image id and placement id will be deleted.
	DeleteID = 'i'
	// Delete newest image with the specified number, specified using the I
	// key. If you specify a p key for the placement id as well, then only the
	// placement with the specified number and placement id will be deleted.
	DeleteNumber = 'n'
	// Delete all placements that intersect with the current cursor position.
	DeleteCursor = 'c'
	// Delete animation frames.
	DeleteFrames = 'f'
	// Delete all placements that intersect a specific cell, the cell is
	// specified using the x and y keys
	DeleteCell = 'p'
	// Delete all placements that intersect a specific cell having a specific
	// z-index. The cell and z-index is specified using the x, y and z keys.
	DeleteCellZ = 'q'
	// Delete all images whose id is greater than or equal to the value of the x
	// key and less than or equal to the value of the y.
	DeleteRange = 'r'
	// Delete all placements that intersect the specified column, specified using
	// the x key.
	DeleteColumn = 'x'
	// Delete all placements that intersect the specified row, specified using
	// the y key.
	DeleteRow = 'y'
	// Delete all placements that have the specified z-index, specified using the
	// z key.
	DeleteZ = 'z'
)
