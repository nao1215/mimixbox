% FAKEMOVIE(1)
% Naohiro CHIKAMATSU <n.chika156@gmail.com>
% September 2021

# NAME

fakemovie –   add fake-movie button to the image.

# SYNOPSIS

**fakemovie** [OPTIONS] IMAGE_FILE_NAME

# DESCRIPTION
**fakemovie** adds a video playback button to the image.  
The supported image formats are png or jpg. Otherwise, an error will occur at runtime.

# EXAMPLES
**When using the default output file name**  
    $ fakemovie image.jpg  
    $ ls  
      image.jpg image_fake.jpg  

**When specifying the output file name, button color, and button size**  
    $ fakemovie -p -o output.jpg -r 50 output.jpg  
    $ ls  
      image.jpg output.jpg  


# OPTIONS
**-o**, **--output**
:   Specify output file name. By default, added suffix "_fake" to original file name.

**-p**, **--phub**
:   Change button color to p-hub. By default, button color is similar to twitter button.

**-r**, **--radius**
:   Specify radius value(integer) of button. The default is an automatically calculated value based on the image size.

**-h**, **--help**
:   Show help message.

**-v**, **--version**
:   Show fakemovie command version.

# EXIT VALUES
**0**
:   Success

**1**
:   There is an error in the argument of the path command or image processing run-time error.

# BUGS
See GitHub Issues: https://github.com/nao1215/mimixbox/issues

# LICENSE
The MimixBox project is licensed under the terms of the MIT license and Apache License 2.0.