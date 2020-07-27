## mille 

mille is a toy text editor in less than 1K lines of code (not including test codes). 
It aims to be simple, readable codes, but works great.
This project is inspired by [https://github.com/antirez/kilo](kilo). 
Moreover, mille doesn't depend on any external library. 

![demo](https://github.com/ad-sho-loko/mille/blob/master/img/demo.gif)

## Features

- Less 1K code bases
- No External Libraries. 
- Implement Gap Buffer

## Editor Features 

- Open file
- Create file
- Save file
- Edit file
- Go syntax highlighting

## Install

If you already installed go, Please type below in your terminal.

```
go get -u github.com/ad-sho-loko/mille
```

## Usage

### Run 

```
mille <filename>
```

### Keys

|  Key  |  Description  |
| ---- | ---- |
|  `Ctrl-H`  |  Backspace |
|  `Ctrl-A`  |  Move Caret to Line Start |
|  `Ctrl-E`  |  Move Caret to Line End |
|  `Ctrl-P`  |  Up |
|  `Ctrl-F`  |  Right |
|  `Ctrl-N`  |  Down |
|  `Ctrl-B`  |  Left |
|  `Ctrl-S`  |  Save |
|  `Ctrl-C`  |  Close |

## Feature works

- UTF8 
- Search/Replace
- Copy/Paste
- Undo/Redo

## Author
Shogo Arakawa (ad.sho.loko@gmail.com)

## LICENSE

MIT
