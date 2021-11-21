# ghrdc - GitHub Release Downloads Counter
ghrdc shows the number of Release file downloads in the repository using GitHub API. By default, it shows the number of downloads for the latest release. Because ghrdc does not authenticate with the GitHub API, it has the following restrictions.   
- It can only be run 60 times per hour.  
- Unable to get information in Organization repository.  

## How to Use
``` Usage
$ ghrdc [OPTIONS] USER_NAME/RPOSITORY_NAME
```

| short option | long option | description |
|:------|:-----|:------|
| -a    | --all    | Show total number of downloads per release.　|
| -t    | --total    |  Show total number of downloads for all releases.　|
| -h   | --help    | Show help message.　 |
| -v | --version  | Show version.|

## Examples
### Get the number of downloads for the latest release
```
$ ghrdc  nao1215/mimixbox
[Name(Version)]             :Version 0.12.1
[Release Date]              :2021-11-20 04:00:23 +0000 UTC
[Binary Download Count]     :0
[Source Code Download Count]:0
```

### Get the total number of downloads for all releases.
```
$ ghrdc -t nao1215/mimixbox
[Name(Version)]                    :All release
[Release Date]                     :-
[Binary Download Count(total)]     :0
[Source Code Download Count(total)]:0
```

### Get the total number of downloads per release.
```
$ ghrdc -a nao1215/mimixbox
[Name(Version)]             :Version 0.12.1
[Release Date]              :2021-11-20 04:00:23 +0000 UTC
[Binary Download Count]     :0
[Source Code Download Count]:0

[Name(Version)]             :Version 0.9.1
[Release Date]              :2021-11-19 07:27:19 +0000 UTC
[Binary Download Count]     :0
[Source Code Download Count]:0

[Name(Version)]             :Version 0.6.0
[Release Date]              :2021-11-18 13:54:57 +0000 UTC
[Binary Download Count]     :0
[Source Code Download Count]:0

[Name(Version)]             :Version 0.3.0
[Release Date]              :2021-11-16 17:17:09 +0000 UTC
[Binary Download Count]     :0
[Source Code Download Count]:0

[Name(Version)]             :Version 0.1.1
[Release Date]              :2021-11-15 16:06:40 +0000 UTC
[Binary Download Count]     :0
[Source Code Download Count]:0

[Name(Version)]             :MimixBox Version 0.0.1 (Buggy)
[Release Date]              :2021-11-14 13:02:27 +0000 UTC
[Binary Download Count]     :0
[Source Code Download Count]:0
```