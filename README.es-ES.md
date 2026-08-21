

## mille 

mille es un editor de texto recreativo con menos de 1K líneas de código (sin incluir el código de pruebas). 
Su objetivo es contar con un código simple y legible, pero funciona de manera excelente.
Este proyecto está inspirado en [https://github.com/antirez/kilo](kilo). 
Además, mille no depende de ninguna biblioteca externa. 

![demo](https://github.com/ad-sho-loko/mille/blob/master/img/demo.gif)

## Características

- Menos de 1K líneas de código
- Sin bibliotecas externas. 
- Implementación de Gap Buffer

## Características del editor 

- Abrir archivo
- Crear archivo
- Guardar archivo
- Editar archivo
- Resaltado de sintaxis de Go

## Instalación

Si ya tienes instalado Go, por favor escribe lo siguiente en tu terminal.

```
go get -u github.com/ad-sho-loko/mille
```

## Uso

### Ejecución 

```
mille <filename>
```

### Teclas

|  Tecla  |  Descripción  |
| ---- | ---- |
|  `Ctrl-H`  |  Retroceso |
|  `Ctrl-A`  |  Mover el cursor al inicio de la línea |
|  `Ctrl-E`  |  Mover el cursor al final de la línea |
|  `Ctrl-P`  |  Arriba |
|  `Ctrl-F`  |  Derecha |
|  `Ctrl-N`  |  Abajo |
|  `Ctrl-B`  |  Izquierda |
|  `Ctrl-S`  |  Guardar |
|  `Ctrl-C`  |  Cerrar |

## Funcionalidades

- UTF8 
- Buscar/Reemplazar
- Copiar/Pegar
- Deshacer/Rehacer

## Autor
Shogo Arakawa (ad.sho.loko@gmail.com)

## LICENCIA

MIT
