# fakemovie: Add movie start button at image
fakemovie add movie start button at image. [Original version made by Mr. mattn](https://github.com/mattn/fakemovie). MimixBox version has some functions added.

## How to Use
``` Usage
fakemovie [OPTIONS] IMAGE_FILE_NAME
```

| short option | long option | description |
|:------|:-----|:------|
|-o| --output| Output file name<br>(default: Added suffix "_fake" to original name) |
|-p|--phub| Put p-hub button<br>(default: Color similar to twitter button)|
|-r|--radius| Radius of button<br>(default: Auto caluculate)|
| -h   | --help    | Show help messages　 |
| -v | --version  | Show version|


## Examples
### Put like Twitter button
```
$ fakemovie lena.png 
```

![Original](../../../images/lena.png "Original")
![Twitter](../../../images/lena_twitter_fake.png "Twitter")

### Put like P-hub button
```
$ fakemovie -p lena.png -o lena_phub_fake.png
```
![Original](../../../images/lena.png "Original")
![Phub](../../../images/lena_phub_fake.png "Phub")