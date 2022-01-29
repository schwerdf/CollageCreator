# CollageCreator

CollageCreator is a Go-language library for the automatic generation
of image collages.

## Library design

As a library, CollageCreator is designed to be extensible by third
parties. Its workflow separates the process of collage creation into
four distinct steps -- reading input images, preprocessing, collage
layout, and output rendering -- and a custom implementation may be
substituted for any of these steps via the API.

### Provided implementations

The core library provides the following implementations for each step:

* _Input image reading_ via Go's
   [built-in image library](https://golang.org/pkg/image/).

* _Preprocessing_ via a command-line switch that lets the user provide
   an [ImageMagick](http://www.imagemagick.org)-like geometry string
   specifying how images are to be scaled and cropped.

* _Collage layout_ via one of two algorithms:

  * _Random placement_: Images are placed at random and then adjusted to
    leave each image equidistant from its nearest neighbor. A binary
    search algorithm is used to minimize the total canvas area.

  * _Tile in order_: Images are placed in rows of identical width,
    images being scaled down to fit. An optimization algorithm is used to
    find a canvas size that minimizes (1) deviation from the provided
    aspect ratio; (2) empty space in the last row or column; and (3) the
    amount by which any image must be scaled down.

* _Output rendering_ as:

  * A PNG, JPEG, or TIFF raster image.

  * A Scalable Vector Graphics (SVG) file, using links to reference
    each input image file.

  * A shell script that runs [ImageMagick](http://www.imagemagick.org)
    tools to build the collage image.

## Dependencies

For raster image output, CollageCreator depends on
[Jan Schlicht's "resize" package](https://github.com/nfnt/resize)
for image scaling.

When the `sh` output file type is selected, the output shell scripts
require the [ImageMagick](http://www.imagemagick.org) command-line
tools to run.

## Known deficiencies

As the built-in Go image library does not read Exif metadata,
CollageCreator does not take Exif orientation into account when
reading images. All images read using this library must be physically
rotated before input to ensure correct operation.
