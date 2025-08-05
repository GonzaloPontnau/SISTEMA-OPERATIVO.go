## **Criterios de Evaluación**

Los grupos deberán concurrir al laboratorio habiendo corrido las pruebas y siendo conscientes de que las mismas funcionan en un entorno distribuido, es decir, **si el trabajo práctico no puede correr en más de una máquina el mismo no se evaluará**.

Al momento de realizar la evaluación en el laboratorio los alumnos dispondrán de un máximo de **10 minutos[^1]** para configurar el ambiente en las computadoras del laboratorio y validar que las conexiones se encuentren funcionando mediante la ejecución de una prueba indicada por el ayudante, caso contrario se considerará que el grupo no se encuentra en condiciones de ser evaluado.

Los grupos contarán con **una única instancia de evaluación por fecha**, es decir, que ante un error no resoluble en el momento, se considerará que el grupo no puede continuar la evaluación y por lo tanto esa entrega se encuentra **desaprobada**, teniendo que presentarse en las siguientes si las hubiera.

---

## **Aclaraciones**

Todos los scripts para realizar las pruebas que se enumeran en este documento se encuentran en la carpeta /scripts
Dentro de las configuraciones propuestas en cada prueba puede haber casos de algunos procesos que no tengan su respectiva configuración porque son valores que no afectan a la prueba en sí.

Los datos de los config que no son provistos en el documento de pruebas es porque dependen de la computadora o del desarrollo de los alumnos (por ejemplo IPs, Puertos o Paths), la configuración del log\_level siempre deberá estar en **INFO**.

Para el proceso IO, ya que no tiene otras configuraciones que no sean IPs y/o Puertos, el archivo de configuración no está detallado en las configs del enunciado de pruebas.

Será responsabilidad del grupo verificar las dependencias requeridas para la compilación, y en caso de requerir bibliotecas provistas por la cátedra, descargarlas e instalarlas en la vm.

Está totalmente prohibido subir archivos binarios o ejecutables al repositorio.

---

## **Prueba Planificación Corto Plazo**

### **Actividades**

1.  Iniciar los módulos.
    1.  Parámetros del Kernel
        1.  archivo\_pseudocodigo: PLANI\_CORTO\_PLAZO
        2.  tamanio\_proceso: 0
2.  Definir el algoritmo de corto plazo a FIFO con dos CPU. Cuando se están ejecutando los procesos infinitos levantar un segundo IO Disco. Luego de verificar que se estén usando los dos IO, matar ambos.
3.  Cambiar el algoritmo de corto plazo a SJF con un solo CPU y volver a ejecutar. Una vez que queden los procesos infinitos matar la IO Disco.
4.  Cambiar el algoritmo de corto plazo a SRT con un solo CPU y volver a ejecutar. Una vez que queden los procesos infinitos matar la IO Disco.

### **Resultados Esperados**

*   Los procesos se ejecutan respetando el algoritmo elegido:
    *   FIFO: Los procesos que no son infinitos terminan sin problemas y los infinitos finalizan luego de matar la IO Disco. Cuando se utilizan dos IO Disco se debe verificar que dos procesos ejecutan su IO en paralelo.
    *   SJF: Se ejecutan los procesos cortos primero y el tiempo de espera promedio del proceso corto es más bajo que el resto.
    *   SRT: Se ejecutan los procesos cortos primero, incluyendo desalojo, el tiempo de espera promedio expresado en las métricas es el más bajo que en SJF.

### **Configuración del sistema**

| *Kernel.config*                                                                                                                  | *CPU1.config*                                                                                                |
| :------------------------------------------------------------------------------------------------------------------------------- | :----------------------------------------------------------------------------------------------------------- |
| ALGORITMO\_CORTO\_PLAZO\=FIFO<br>ALGORITMO\_INGRESO\_A\_READY\=FIFO<br>ALFA\=1<br>ESTIMACION\_INICIAL\=10000<br>TIEMPO\_SUSPENSION\=120000 | ENTRADAS\_TLB\=4<br>REEMPLAZO\_TLB\=LRU<br>ENTRADAS\_CACHE\=2<br>REEMPLAZO\_CACHE\=CLOCK<br>RETARDO\_CACHE\=250 |
| *Memoria.config*                                                                                                                 | *CPU2.config*                                                                                                |
| TAM\_MEMORIA\=4096<br>TAM\_PAGINA\=64<br>ENTRADAS\_POR\_TABLA\=4<br>CANTIDAD\_NIVELES\=2<br>RETARDO\_MEMORIA\=500<br>RETARDO\_SWAP\=15000  | ENTRADAS\_TLB\=4<br>REEMPLAZO\_TLB\=LRU<br>ENTRADAS\_CACHE\=2<br>REEMPLAZO\_CACHE\=CLOCK<br>RETARDO\_CACHE\=250  |
| *Lista de Procesos IO*                                                                                                           |                                                                                                              |
| DISCO (1 instancia)                                                                                                              |                                                                                                              |

---

## **Prueba Planificación Mediano/Largo Plazo**

### **Actividades**

1.  Iniciar los módulos.
    1.  Parámetros del Kernel
        1.  archivo\_pseudocodigo: PLANI\_LYM\_PLAZO
        2.  tamanio\_proceso: 0
2.  Esperar la finalización de los mismos con normalidad.
3.  Cambiar el algoritmo de ingreso a ready a PMCP y volver a ejecutar.

### **Resultados Esperados**

*   Los procesos finalizan con normalidad.
*   Todos los procesos PLANI\_LYM\_IO deben ser suspendidos y lo que se encuentra en la memoria de dicho proceso debe ser enviado a SWAP.
*   Para PMCP el último proceso en ingresar a NEW debe ser el que solicita 256 de memoria.

### **Configuración del sistema**

| *Kernel.config*                                                                                                            | *CPU.config*                                                                                                 |
| :------------------------------------------------------------------------------------------------------------------------- | :----------------------------------------------------------------------------------------------------------- |
| ALGORITMO\_CORTO\_PLAZO\=FIFO<br>ALGORITMO\_INGRESO\_A\_READY\=FIFO<br>ALFA\=1<br>ESTIMACION\_INICIAL\=10000<br>TIEMPO\_SUSPENSION\=3000 | ENTRADAS\_TLB\=4<br>REEMPLAZO\_TLB\=LRU<br>ENTRADAS\_CACHE\=2<br>REEMPLAZO\_CACHE\=CLOCK<br>RETARDO\_CACHE\=250 |
| *Memoria.config*                                                                                                           | *Lista de Procesos IO*                                                                                       |
| TAM\_MEMORIA\=256<br>TAM\_PAGINA\=16<br>ENTRADAS\_POR\_TABLA\=4<br>CANTIDAD\_NIVELES\=2<br>RETARDO\_MEMORIA\=500<br>RETARDO\_SWAP\=3000    | DISCO (1 instancia)                                                                                          |

---

## **Prueba Memoria SWAP**

### **Actividades**

1.  Iniciar los módulos.
    1.  Parámetros del Kernel:
        1.  archivo\_pseudocodigo: MEMORIA\_IO
        2.  tamanio\_proceso: 90
2.  Esperar a que los procesos finalicen.

### **Resultados Esperados**

Se puede observar el contenido de la memoria en los archivos de DUMP o en el archivo de SWAP y en ambos casos son valores consistentes.

### **Configuración del sistema**

| *Kernel.config*                                                                                                            | *CPU1.config*                                                                                                |
| :------------------------------------------------------------------------------------------------------------------------- | :----------------------------------------------------------------------------------------------------------- |
| ALGORITMO\_CORTO\_PLAZO\=FIFO<br>ALGORITMO\_INGRESO\_A\_READY\=FIFO<br>ALFA\=1<br>ESTIMACION\_INICIAL\=10000<br>TIEMPO\_SUSPENSION\=1000 | ENTRADAS\_TLB\=0<br>REEMPLAZO\_TLB\=FIFO<br>ENTRADAS\_CACHE\=0<br>REEMPLAZO\_CACHE\=CLOCK<br>RETARDO\_CACHE\=250 |
| *Memoria.config*                                                                                                           | *Lista de Procesos IO*                                                                                       |
| TAM\_MEMORIA\=512<br>TAM\_PAGINA\=32<br>ENTRADAS\_POR\_TABLA\=32<br>CANTIDAD\_NIVELES\=1<br>RETARDO\_MEMORIA\=500<br>RETARDO\_SWAP\=2500   | DISCO                                                                                                        |

---

## **Prueba Memoria - Caché**

### **Actividades**

1.  Iniciar los módulos.
    1.  Parámetros del Kernel:
        1.  archivo\_pseudocodigo: MEMORIA\_BASE
        2.  tamanio\_proceso: 256
2.  Esperar a que el proceso finalice.
3.  Cambiar el algoritmo de reemplazo de caché a CLOCK-M y volver a lanzar la prueba
4.  Esperar que el proceso finalice.

### **Resultados Esperados**

Los reemplazos de caché se realizan de acuerdo a los algoritmos definidos.

### **Configuración del sistema**

| *Kernel.config*                                                                                                            | *CPU1.config*                                                                                                |
| :------------------------------------------------------------------------------------------------------------------------- | :----------------------------------------------------------------------------------------------------------- |
| ALGORITMO\_CORTO\_PLAZO\=FIFO<br>ALGORITMO\_INGRESO\_A\_READY\=FIFO<br>ALFA\=1<br>ESTIMACION\_INICIAL\=10000<br>TIEMPO\_SUSPENSION\=3000 | ENTRADAS\_TLB\=0<br>REEMPLAZO\_TLB\=FIFO<br>ENTRADAS\_CACHE\=2<br>REEMPLAZO\_CACHE\=CLOCK<br>RETARDO\_CACHE\=250 |
| *Memoria.config*                                                                                                           | *Lista de Procesos IO*                                                                                       |
| TAM\_MEMORIA\=2048<br>TAM\_PAGINA\=32<br>ENTRADAS\_POR\_TABLA\=4<br>CANTIDAD\_NIVELES\=3<br>RETARDO\_MEMORIA\=500<br>RETARDO\_SWAP\=5000   | DISCO                                                                                                        |

---

## **Prueba Memoria - TLB**

### **Actividades**

1.  Iniciar los módulos.
    1.  Parámetros del Kernel:
        1.  archivo\_pseudocodigo: MEMORIA\_BASE
        2.  tamanio\_proceso: 256
2.  Esperar a que el proceso finalice.
3.  Cambiar el algoritmo de reemplazo de TLB a LRU y volver a lanzar la prueba
4.  Esperar que el proceso finalice.

### **Resultados Esperados**

Los reemplazos de TLB se dan de acuerdo a los algoritmos definidos.

### **Configuración del sistema**

| *Kernel.config*                                                                                                            | *CPU1.config*                                                                                                |
| :------------------------------------------------------------------------------------------------------------------------- | :----------------------------------------------------------------------------------------------------------- |
| ALGORITMO\_CORTO\_PLAZO\=FIFO<br>ALGORITMO\_INGRESO\_A\_READY\=FIFO<br>ALFA\=1<br>ESTIMACION\_INICIAL\=10000<br>TIEMPO\_SUSPENSION\=3000 | ENTRADAS\_TLB\=4<br>REEMPLAZO\_TLB\=FIFO<br>ENTRADAS\_CACHE\=0<br>REEMPLAZO\_CACHE\=CLOCK<br>RETARDO\_CACHE\=250 |
| *Memoria.config*                                                                                                           | *Lista de Procesos IO*                                                                                       |
| TAM\_MEMORIA\=2048<br>TAM\_PAGINA\=32<br>ENTRADAS\_POR\_TABLA\=4<br>CANTIDAD\_NIVELES\=3<br>RETARDO\_MEMORIA\=500<br>RETARDO\_SWAP\=5000   | DISCO                                                                                                        |

---

## **Prueba Estabilidad General**

### **Actividades**

1.  Iniciar los módulos.
    1.  Parámetros del Kernel:
        1.  archivo\_pseudocodigo: ESTABILIDAD\_GENERAL
        2.  tamanio\_proceso: 0
2.  Dejar correr todo por un buen tiempo.

### **Resultados Esperados**

No se observan esperas activas ni memory leaks y el sistema no finaliza de manera abrupta.

### **Configuración del sistema**

| *Kernel.config*                                                                                                                  | *Memoria.config*                                                                                                         |
| :------------------------------------------------------------------------------------------------------------------------------- | :----------------------------------------------------------------------------------------------------------------------- |
| ALGORITMO\_CORTO\_PLAZO\=SRT<br>ALGORITMO\_INGRESO\_A\_READY\=PMCP<br>ALFA\=0.75<br>ESTIMACION\_INICIAL\=100<br>TIEMPO\_SUSPENSION\=3000 | TAM\_MEMORIA\=4096<br>TAM\_PAGINA\=32<br>ENTRADAS\_POR\_TABLA\=8<br>CANTIDAD\_NIVELES\=3<br>RETARDO\_MEMORIA\=100<br>RETARDO\_SWAP\=2500 |
| *CPU1.config*                                                                                                                    | *CPU2.config*                                                                                                            |
| ENTRADAS\_TLB\=4<br>REEMPLAZO\_TLB\=FIFO<br>ENTRADAS\_CACHE\=2<br>REEMPLAZO\_CACHE\=CLOCK<br>RETARDO\_CACHE\=50                       | ENTRADAS\_TLB\=4<br>REEMPLAZO\_TLB\=LRU<br>ENTRADAS\_CACHE\=2<br>REEMPLAZO\_CACHE\=CLOCK-M<br>RETARDO\_CACHE\=50              |
| *CPU3.config*                                                                                                                    | *CPU4.config*                                                                                                            |
| ENTRADAS\_TLB\=256<br>REEMPLAZO\_TLB\=FIFO<br>ENTRADAS\_CACHE\=256<br>REEMPLAZO\_CACHE\=CLOCK<br>RETARDO\_CACHE\=1                    | ENTRADAS\_TLB\=0<br>REEMPLAZO\_TLB\=FIFO<br>ENTRADAS\_CACHE\=0<br>REEMPLAZO\_CACHE\=CLOCK<br>RETARDO\_CACHE\=0                 |
| *Lista de Procesos IO*                                                                                                           |                                                                                                                          |
| DISCO (4 instancias)                                                                                                             |                                                                                                                          |

---
---

