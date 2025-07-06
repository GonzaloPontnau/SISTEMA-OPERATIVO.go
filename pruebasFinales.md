### Criterios de Evaluación
Los grupos deberán concurrir al laboratorio habiendo corrido las pruebas y siendo conscientes de que las mismas funcionan en un entorno distribuido, es decir, si el trabajo práctico no puede correr en más de una máquina el mismo no se evaluará.
Al momento de realizar la evaluación en el laboratorio los alumnos dispondrán de un máximo de 10 minutos para configurar el ambiente en las computadoras del laboratorio y validar que las conexiones se encuentren funcionando mediante la ejecución de una prueba indicada por el ayudante, caso contrario se considerará que el grupo no se encuentra en condiciones de ser evaluado.
Los grupos contarán con una única instancia de evaluación por fecha, es decir, que ante un error no resoluble en el momento, se considerará que el grupo no puede continuar la evaluación y por lo tanto esa entrega se encuentra desaprobada, teniendo que presentarse en las siguientes si las hubiera.

### Aclaraciones
Todos los scripts para realizar las pruebas que se enumeran en este documento se encuentran la carpeta /scripts
Dentro de las configuraciones propuestas en cada prueba puede haber casos de algunos procesos que no tengan su respectiva configuración porque son valores que no afectan a la prueba en sí.
Los datos de los config que no son provistos en el documento de pruebas es porque dependen de la computadora o del desarrollo de los alumnos (por ejemplo IPs, Puertos o Paths), la configuración del log_level siempre deberá estar en INFO.
Para el proceso IO, ya que no tiene otras configuraciones que no sean IPs Y/o Puertos, el archivo de configuración no está detallado en las configs del enunciado de pruebas.
Será responsabilidad del grupo verificar las dependencias requeridas para la compilación, y en caso de requerir bibliotecas provistas por la cátedra, descargarlas e instalarlas en la vm.
Está totalmente prohibido subir archivos binarios al repositorio.
1 Recomendamos leer la Guía de Deploy

---

### Prueba Planificación Corto Plazo

#### Actividades
1. Iniciar los módulos.
   Parámetros del Kernel
   a.
   i. archivo_pseudocodigo: PLANI_CORTO_PLAZO
   ii. tamanio_proceso: 0
2. Definir el algoritmo de corto plazo a FIFO con dos CPU. Cuando se están ejecutando los procesos infinitos levantar un segundo IO Disco. Luego de verificar que se estén usando los dos IO, matar ambos.
3. Cambiar el algoritmo de corto plazo a SJF con un solo CPU y volver a ejecutar. Una vez que queden los procesos infinitos matar la IO Disco.
4. Cambiar el algoritmo de corto plazo a SRT con un solo CPU y volver a ejecutar. Una vez que queden los procesos infinitos matar la IO Disco.

#### Resultados Esperados
Los procesos se ejecutan respetando el algoritmo elegido:
FIFO: Los procesos que no son infinitos terminan sin problemas y los infinitos finalizan luego de matar la IO Disco. Cuando se utilizan dos IO Disco se debe verificar que dos procesos ejecutan su IO en paralelo.
SJF: Se ejecutan los procesos cortos primero y el tiempo de espera promedio en ready es más bajo que en FIFO.
SRT: Se ejecutan los procesos cortos primero, incluyendo desalojo, el tiempo de espera promedio expresado en las métricas es el más bajo.

#### Configuración del sistema

| Kernel.config | CPU1.config |
|---|---|
| ALGORITMO_CORTO_PLAZO=FIFO | ENTRADAS_TLB =4 |
| ALGORITMO_INGRESO_A_READY =FIF0 | REEMPLAZO_TLB =LRU |
| ALFA =1 | ENTRADAS_CACHE =2 |
| ESTIMACION_INICIAL =1000 | REEMPLAZO_CACHE =CLOCK |
| TIEMPO_SUSPENSION=120000 | RETARDO_CACHE =250 |
| **Memoria.config** | **CPU2.config** |
| TAM_MEMORIA =4096 | ENTRADAS_TLB =4 |
| TAM_PAGINA =64 | REEMPLAZO_TLB =LRU |
| ENTRADAS_POR_TABLA =4 | ENTRADAS_CACHE =2 |
| CANTIDAD_NIVELES =2 | REEMPLAZO_CACHE=CLOCK |
| RETARDO MEMORIA =500 | RETARDO_CACHE =250 |
| RETARDO_SWAP=15000 | |

---

Lista de Procesos 10
DISCO (1 instancia)

---

### Prueba Planificación Mediano/Largo Plazo

#### Actividades
1. Iniciar los módulos.
   a. Parámetros del Kernel
      i. archivo_pseudocodigo: PLANI_LYM_PLAZO
      ii. tamanio_proceso: 0
2. Esperar la finalización de los mismos con normalidad.
3. Cambiar el algoritmo de ingreso a ready a PMCP y volver a ejecutar.

#### Resultados Esperados
Los procesos finalizan con normalidad.
Todos los procesos PLANI_LYM_IO deben ser suspendidos y lo que se encuentra en la memoria de dicho proceso debe ser enviado a SWAP.
• Para PMCP el último proceso en ingresar a NEW debe ser el que solicita 256 de memoria.

#### Configuración del sistema

| Kernel.config | CPU.config |
|---|---|
| ALGORITMO_CORTO_PLAZO=FIFO<br>ALGORITMO_INGRESO_A_READY=FIFO<br>ALFA =1<br>ESTIMACION_INICIAL =10000<br>TIEMPO_SUSPENSION =3000 | ENTRADAS_TLB =4<br>REEMPLAZO_TLB =LRU<br>ENTRADAS_CACHE =2<br>REEMPLAZO_CACHE =CLOCK<br>RETARDO_CACHE =250 |
| **Memoria.config** | **Lista de Procesos 10** |
| TAM_MEMORIA =256 | DISCO (1 instancia) |
| TAM_PAGINA =16<br>ENTRADAS_POR_TABLA =4<br>CANTIDAD_NIVELES =2<br>RETARDO_MEMORIA =500<br>RETARDO_SWAP =3000 | |

---

### Prueba Memoria

#### Actividades
1. Iniciar los módulos.
   Parámetros del Kernel:
   a.
   i. archivo_pseudocodigo: MEMORIA
   ii. tamanio_proceso: 0
2. Esperar a que los procesos creados se queden en estado SUSP. BLOCKED.

#### Resultados Esperados
En las CPU 1 y 2 los reemplazos de caché se realizan de acuerdo a los algoritmos definidos
En la CPU 3 todos los accesos se realizan directamente sobre memoria.

#### Configuración del sistema

| Kernel.config | CPU1.config |
|---|---|
| ALGORITMO_CORTO_PLAZO=FIFO | ENTRADAS_TLB =4 |
| ALGORITMO_INGRESO_A_READY =FIFO<br>ALFA =1<br>ESTIMACION_INICIAL =1000 | REEMPLAZO_TLB =FIFO<br>ENTRADAS_CACHE =2<br>REEMPLAZO_CACHE ==CLOCK |
| TIEMPO_SUSPENSION =3000 | RETARDO_CACHE =250 |
| **Memoria.config** | **CPU2.config** |
| TAM_MEMORIA =2048 | ENTRADAS_TLB =4 |
| TAM_PAGINA =32 | REEMPLAZO_TLB =LRU |
| ENTRADAS_POR_TABLA =4 | ENTRADAS_CACHE =2 |
| CANTIDAD_NIVELES =3<br>RETARDO_MEMORIA =500<br>RETARDO_SWAP =500 | REEMPLAZO_CACHE=CLOCK-M<br>RETARDO CACHE =250 |
| **Lista de Procesos IO** | **CPU3.config** |
| DISCO (3 instancias) | ENTRADAS_TLB =0<br>REEMPLAZO_TLB=FIFO |
| | ENTRADAS_CACHE =0 |
| | REEMPLAZO_CACHE=CLOCK |
| | RETARDO_CACHE =0 |

---

### Prueba Estabilidad General

#### Actividades
1. Iniciar los módulos.
   Parámetros del Kernel:
   a.
   i. archivo_pseudocodigo: ESTABILIDAD_GENERAL
   ii. tamanio_proceso: 0
2. Dejar correr todo por un buen tiempo

#### Resultados Esperados
No se observan esperas activas ni memory leaks y el sistema no finaliza de manera abrupta.

#### Configuración del sistema

| Kernel.config | Memoria.config |
|---|---|
| ALGORITMO_CORTO_PLAZO =SRT<br>ALGORITMO_INGRESO_A_READY =PMCP<br>ALFA =0.75 | TAM_MEMORIA=4096<br>TAM_PAGINA =32<br>ENTRADAS_POR_TABLA =8 |
| ESTIMACION_INICIAL =100 | CANTIDAD_NIVELES =3 |
| TIEMPO_SUSPENSION =3000 | RETARDO_MEMORIA =100<br>RETARDO_SWAP =2500 |
| **CPU1.config** | **CPU2.config** |
| ENTRADAS_TLB =4 | ENTRADAS_TLB =4 |
| REEMPLAZO_TLB =FIFO | REEMPLAZO_TLB =LRU |
| ENTRADAS_CACHE =2<br>REEMPLAZO_CACHE =CLOCK<br>RETARDO_CACHE =50 | ENTRADAS_CACHE =2<br>REEMPLAZO_CACHE =CLOCK-M<br>RETARDO_CACHE =50 |
| **CPU3.config** | **CPU4.config** |
| ENTRADAS_TLB =256 | ENTRADAS_TLB =0 |
| REEMPLAZO_TLB =FIFO | REEMPLAZO_TLB =FIFO |
| ENTRADAS_CACHE =256 | ENTRADAS_CACHE =0 |
| REEMPLAZO_CACHE=CLOCK<br>RETARDO_CACHE =1 | REEMPLAZO_CACHE=CLOCK<br>RETARDO_CACHE =0 |
| **Lista de Procesos 10** | |
| DISCO (4 instancias) | |

---

### Planilla de Evaluación - TP1C2025

| Nombre del Grupo | Nota (Grupal) |
|---|---|
| | |

| Legajo | Apellido y Nombres | Nota (Coloquio) |
|---|---|---|
| | | |
| | | |
| | | |

| Evaluador/es Práctica | Evaluador/es Coloquio |
|---|---|
| | |

Observaciones:

---

| Sistema Completo | |
|---|---|
| El deploy se hace compilando los módulos en las máquinas del laboratorio en menos de 10 minutos. | |
| Los procesos se ejecutan de forma simultánea y la cantidad de hilos y subprocesos en el sistema es la adecuada. | |
| Los procesos establecen conexiones TCP/IP. | |
| El sistema no registra casos de Espera Activa ni Memory Leaks. | |
| El log respeta los lineamientos de logs mínimos y obligatorios de cada módulo | |
| El sistema no requiere permisos de superuser (sudo/root) para ejecutar correctamente. | |
| El sistema no requiere de Valgrind o algún proceso similar para ejecutar correctamente. | |
| El sistema utiliza una sincronización determinística (no utiliza más sleeps de los solicitados). | |

| Módulo Kernel | |
|---|---|
| Interpreta correctamente los comandos introducidos por su consola. | |
| Respeta el grado de multiprogramación definido por archivo de configuración. | |
| Se respeta el diagrama de 5 estados y sus transiciones. | |
| El planificador de corto plazo respeta el orden de llegada de los procesos en FIFO. | |
| El planificador de corto plazo ejecuta correctamente las syscalls. | |
| El planificador de corto plazo respeta las estimaciones en SJF y SRT. | |
| El planificador de corto plazo envía las interrupciones a la CPU para desalojar procesos.. | |
| Se permite la conexión y desconexión de instancias de IO sin presentar errores. | |

| Módulo CPU | |
|---|---|
| Respeta el ciclo de instrucción. | |
| Actualiza correctamente el PC antes de devolverlo al kernel. | |
| Interpreta correctamente las instrucciones definidas. | |
| Realiza las traducciones de dirección lógica a física siguiendo lo definido en el enunciado. | |
| Los accesos a memoria se realizan correctamente. | |
| La implementación de TLB respeta los límites definidos | |
| La implementación de la TLB respeta y ejecuta correctamente los algoritmos de reemplazo. | |
| La implementación de la caché de páginas respeta los límites definidos | |
| La implementación de la caché de páginas respeta y ejecuta correctamente los algoritmos de reemplazo | |

---

| Módulo Memoria | |
|---|---|
| Se respetan los tamaños de página. | |
| Se respetan los retardos en las operaciones. | |
| Se administra correctamente el espacio de usuario utilizando un único void* o slice | |
| Permite la creación y finalización de procesos. | |
| Permite acceder correctamente a las tablas de páginas. | |
| Permite acceder al espacio de usuario únicamente a través de direcciones físicas. | |
| Permite la suspensión y desuspensión de procesos. | |
| Maneja el espacio de SWAP como un único archivo. | |