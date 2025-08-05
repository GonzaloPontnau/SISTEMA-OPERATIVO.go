## Episodio IX

### The Rise of Gopher

---

## Arquitectura del sistema

Este trabajo práctico correrá una serie de módulos los cuales deberán ser capaces de correr en diferentes computadoras y/o máquinas virtuales.

> **NOTA:** El sentido de las flechas indica dependencias entre módulos. Ejemplo: al momento de iniciar la ejecución del Kernel es necesario contar con la Memoria + SWAP iniciada.

(Diagrama: Kernel conectado a CPU y Memoria + SWAP, IO también conectado a Kernel)

*Poner imagen*

## Módulos

Todos los módulos de la arquitectura serán API's que se deberán exponer a través del protocolo HTTP. Esta arquitectura nos permitirá una comunicación en ambos sentidos pudiendo modularizar y atomizar las operaciones que se realicen.

De esta manera, cada acción de “Módulo A debe comunicar a Módulo B" define una API que debe ser creada en el Módulo B mientras que cada acción de “Módulo B debe comunicar a Módulo A" define una API que debe ser creada en el Módulo A.

## Aclaración importante

Desarrollar únicamente temas de **conectividad, serialización, sincronización o el módulo IO** es insuficiente para poder entender y aprender los distintos conceptos de la materia. Dicho caso será un motivo de **desaprobación directa**.

Cada módulo contará con un listado de **logs mínimos y obligatorios** los cuales deberán realizarse utilizando la biblioteca de `log/slog` sugerida por la cátedra y los mismos deberán estar como `Logger.Info()`, pudiendo ser extendidos por necesidad del grupo utilizando `Logger.Debug()`.

---

En caso de no cumplir con los logs mínimos y/o no persistirlos en archivo, **se considerará que el TP no es apto para ser evaluado y por consecuencia el mismo estará desaprobado.**

---

## Módulo: Kernel

El módulo Kernel, en el contexto de nuestro trabajo práctico, será el módulo encargado de la gestión de los procesos que se creen dentro de este sistema simulado, planificando su ejecución en múltiples CPUs mediante diferentes algoritmos, llevando a cabo sus peticiones de entrada/salida hacia múltiples dispositivos (representados por el módulo I/O) e interactuando con el módulo Memoria para su utilización.

### Lineamiento e Implementación

Para lograr su propósito, este módulo mantendrá conexiones efímeras vía API's con una o más instancias del módulo CPU y con el módulo Memoria.

* Con CPU: consumir mínimamente 2 APIs:

  1. Envío de procesos a ejecutar y espera de respuesta en API expuesta.
  2. Notificaciones de interrupciones para algoritmos preemtivos.
* Con Memoria: invocar API's que respondan de forma sincrónica.

Al iniciar el módulo, se creará un proceso inicial para que este lo planifique y para poder inicializarlo, se requerirá que este módulo reciba dos argumentos adicionales en la función `main`: el nombre del archivo de pseudocódigo que deberá ejecutar y el tamaño del proceso para ser inicializado en Memoria.

```bash
~ ./bin/kernel [archivo_pseudocodigo] [tamanio_proceso] [...args]
```

### Diagrama de estados

El kernel utilizará un diagrama de **7 estados** para la planificación de los procesos e hilos:

```text
NEW → READY → EXEC → BLOCKED → EXIT
  ↑           ↖         ↗
 SUSP.READY  SUSP.BLOCKED
```

*(Poner diagrama con los 7 estados)*

---

## PCB

El PCB será la estructura base que utilizaremos dentro del Kernel para administrar cada uno de los procesos. Deberá contener como mínimo:

* **PID:** Identificador entero único, arranca en 0.
* **PC:** Program Counter, entero que arranca en 0.
* **ME:** Métricas de Estado, conteo de veces en cada estado.
* **MT:** Métricas de Tiempo, tiempo en ms en cada estado.

## Planificador de Largo Plazo

El Kernel gestionará las peticiones a la Memoria para creación y eliminación de procesos.

### Inicio

Al iniciar, el algoritmo de Largo Plazo estará detenido (`STOP`) y esperará un `Enter` por teclado para comenzar.

### Creación de procesos

Se tendrá una cola `NEW` administrada por el algoritmo definido en config (`FIFO` o `Proceso más chico primero`).

* Si `NEW` y `SUSP.READY` están vacías: pedir a Memoria inicializar proceso.

  * Respuesta positiva → pasar a `READY`.
  * Respuesta negativa → esperar finalización de otro proceso.
* Si hay procesos esperando:

  * **FIFO:** el primero bloqueado impide a los siguientes.
  * **PMCP:** cada llegada consulta Memoria; si hay espacio, entra; sino, se encola. A liberar espacio, ordenar por tamaño asc.

#### FIFO

Bajo FIFO, si el primer proceso no puede iniciar por falta de espacio, los siguientes esperan.

#### Proceso más chico primero

Prioriza el proceso de menor tamaño:

* Al llegar, consulta Memoria.
* Si entra → `READY`.
* Si no → en cola hasta espacio.
* Al liberar, revalida cola ordenada asc.

### Finalización de procesos

Al finalizar, informar a Memoria y, tras confirmación:

1. Liberar PCB.
2. Intentar inicializar procesos en `SUSP.READY` primero, luego `NEW`.
3. Loguear métricas de estado:

```text
## (<PID>) - Métricas de estado: NEW (count) (time), READY (count) (time), ...
```

## Planificador de Corto Plazo

Los procesos en `READY` se planifican por algoritmo config:

* **FIFO**
* **SJF sin Desalojo**
* **SJF con Desalojo (SRT)**

### FIFO

Elige según orden de llegada a `READY`.

### SJF sin Desalojo

Elige rafaga más corta. Estimación:

```text
Est(n): estimado anterior
R(n): ráfaga real anterior
Est(n+1) = α·R(n) + (1−α)·Est(n)
α ∈ [0,1]
```

### SJF con Desalojo (SRT)

Si nuevo proceso en `READY` tiene ráfaga menor que alguno en ejecución, notificar CPU con mayor restante para desalojar.

## Ejecución

1. Pasar proceso a `EXEC` y enviar a CPU:

```text
dispatch → { PID, PC }
```

2. Esperar retorno `{ PID, PC, motivo }`.
3. Si motivo requiere replanificar, seleccionar siguiente.

Para desalojo, usar endpoint de `interrupt`.

## Syscalls

### Procesos

* **INIT\_PROC**(archivo, tamaño): Kernel crea PCB en `NEW`. No cambia estado del proceso invocante.
* **EXIT**(): Finaliza proceso invocante.

### Memoria

* **DUMP\_MEMORY**: Solicita dump; bloquea hasta respuesta. Error → `EXIT`; éxito → `READY`.

## Entrada/Salida

El Kernel gestiona módulos IO conectados por API:

* Conocer IP, PUERTO, procesos en I/O y en cola.
* Syscall **IO**(nombre, ms):

  1. Validar existencia del dispositivo.
  2. Si no existe → `EXIT`.
  3. Si existe:

     * Si libre → enviar `{ PID, tiempo }` a IO.
     * Si ocupado → `BLOCKED` y en cola.
* Al desconexión de IO: proceso en ejecución pasa a `EXIT`; procesos en cola si no quedan instancias → `EXIT`.
* Al fin de I/O: desbloquear siguiente o pasar a `READY`.

## Planificador de Mediano Plazo

Al bloquear (`BLOCKED`), se inicia timer (configurable);

* Al expirar, si sigue `BLOCKED` → `SUSP.BLOCKED`; informar a Memoria para swap.
* Al liberar espacio, intentar mover de `NEW`/`SUSP.READY`.
* Tras fin de I/O en swap → `SUSP.READY`
* Para pasar a `READY`, siempre priorizar `SUSP.READY` sobre `NEW`.

> ^1 Asegurar handlers para señales y timeout de APIs.

## Consideraciones teóricas

Los esquemas son simplificaciones académicas y no reflejan necesariamente sistemas reales.

## Logs mínimos y obligatorios

```text
## (<PID>) - Solicitó syscall: <NOMBRE_SYSCALL>
## (<PID>) - Se crea el proceso - Estado: NEW
## (<PID>) - Pasa del estado <ANTERIOR> al estado <ACTUAL>
## (<PID>) - Bloqueado por IO: <DISPOSITIVO>
## (<PID>) - Finalizó IO y pasa a READY
## (<PID>) - Desalojado por algoritmo SJF/SRT
## (<PID>) - Finaliza el proceso
## (<PID>) - Métricas de estado: NEW (cnt)(time), READY (cnt)(time), ...
```

## Archivo de Configuración

| Campo                     | Tipo     | Descripción                              |
| :------------------------ | :------- | :--------------------------------------- |
| ip\_memory                | String   | IP de la Memoria                         |
| port\_memory              | Numérico | Puerto de la Memoria                     |
| scheduler\_algorithm      | String   | FIFO/SJF/SRT                             |
| ready\_ingress\_algorithm | String   | FIFO/PMCP                                |
| alpha                     | Numérico | Valor α para estimación SJF              |
| initial\_estimate         | Numérico | Estimación inicial de ráfaga             |
| suspension\_time          | Numérico | Tiempo antes de suspender a SUSP.BLOCKED |
| log\_level                | String   | Nivel de logs con `slog.SetLogLevel()`   |
| port\_kernel              | Numérico | Puerto del Kernel                        |
| ip\_kernel                | String   | IP del Kernel                            |

### Ejemplo de Configuración

```json
{
  "ip_memory": "127.0.0.1",
  "port_memory": 8002,
  "ip_kernel": "127.0.0.1",
  "port_kernel": 8001,
  "scheduler_algorithm": "FIFO",
  "ready_ingress_algorithm": "PMCP",
  "alpha": 0.5,
  "initial_estimate": 10000,
  "suspension_time": 4500,
  "log_level": "DEBUG"
}
```

---

## Módulo: IO

El módulo IO simula dispositivos de Entrada/Salida.

### Lineamiento e Implementación

1. Iniciar con:

   ```bash
   ~ ./bin/io [nombre]
   ```
2. Handshake con Kernel: enviar `nombre`, IP y Puerto.
3. Al recibir petición: `usleep(tiempo)`.
4. Finalizar petición → notificar Kernel.

### Finalización del Módulo IO

Manejar `SIGINT` y `SIGTERM` para notificar y cerrar controladamente.

### Logs mínimos y obligatorios

```text
## PID: <PID> - Inicio de IO - Tiempo: <TIEMPO_IO>
## PID: <PID> - Fin de IO
```

### Archivo de Configuración

| Campo        | Tipo     | Descripción                          |
| :----------- | :------- | :----------------------------------- |
| port\_io     | Numérico | Puerto del módulo IO                 |
| ip\_io       | String   | IP del módulo IO                     |
| ip\_kernel   | String   | IP del Kernel                        |
| port\_kernel | Numérico | Puerto del Kernel                    |
| log\_level   | String   | Nivel de logs (`slog.SetLogLevel()`) |

### Ejemplo

```json
{
  "ip_kernel": "127.0.0.1",
  "port_kernel": 8001,
  "port_io": 8003,
  "ip_io": "127.0.0.1",
  "log_level": "DEBUG"
}
```

---

## Módulo: CPU

Simula el ciclo de instrucción de una CPU.

### Lineamiento e Implementación

1. Iniciar con:

   ```bash
   ~ /bin/cpu [identificador]
   ```
2. Handshake con Kernel: enviar IP, Puerto e identificador.
3. Al recibir `{ PID, PC }`:

   * Pedir instrucción a Memoria.
   * Ejecutar ciclo: **Fetch → Decode → Execute → Check Interrupt**.

### Ciclo de Instrucción

* **Fetch:** solicitar instrucción a Memoria usando PC.
* **Decode:** interpretar instrucción y necesidad de traducir direcciones.
* **Execute:** según tipo:

  * `NOOP`: consumir ciclo.
  * `WRITE dir datos`: escribir cadena (sin espacios).
  * `READ dir tamaño`: leer y loguear.
  * `GOTO valor`: actualizar PC.
  * Syscalls: `IO`, `INIT_PROC`, `DUMP_MEMORY`, `EXIT`.
* **Check Interrupt:** si llegó interrupción al PID, retornar con motivo.

#### Instrucciones de ejemplo

```text
NOOP
WRITE 0 EJEMPLO_DE_ENUNCIADO
READ 0 20
GOTO 0
IO IMPRESORA 25000
INIT_PROC proceso1 256
DUMP_MEMORY
EXIT
```

### MMU, TLB y Caché

Ver sección `Memoria + SWAP` para detalles de paginación multinivel, TLB y caché.

### Logs mínimos y obligatorios

```text
## PID: <PID> - FETCH - PC: <PC>
## Interrupción recibida al puerto Interrupt
## PID: <PID> - Ejecutando: <INSTRUCCION> <ARGS>
## PID: <PID> - Acción: <LEER/ESCRIBIR> - Dir Física: <DIR> - Valor: <VALOR>
## PID: <PID> - OBTENER MARCO - Página: <PAG> - Marco: <MAR>
## PID: <PID> - TLB HIT/MISS - Página: <PAG>
## PID: <PID> - Cache Hit/Miss/Add - Página: <PAG>
## PID: <PID> - Memory Update - Página: <PAG> - Frame: <FRAME>
```

### Archivo de Configuración

| Campo              | Tipo     | Descripción                                 |
| :----------------- | :------- | :------------------------------------------ |
| port\_cpu          | Numérico | Puerto del CPU                              |
| ip\_cpu            | String   | IP del CPU                                  |
| ip\_memory         | String   | IP de la Memoria                            |
| port\_memory       | Numérico | Puerto de la Memoria                        |
| ip\_kernel         | String   | IP del Kernel                               |
| port\_kernel       | Numérico | Puerto del Kernel                           |
| tlb\_entries       | Numérico | Cantidad de entradas de la TLB              |
| tlb\_replacement   | String   | FIFO o LRU                                  |
| cache\_entries     | Numérico | Cantidad de entradas de la caché de páginas |
| cache\_replacement | String   | CLOCK o CLOCK-M                             |
| cache\_delay       | Numérico | Delay antes de responder                    |
| log\_level         | String   | Nivel de logs                               |

### Ejemplo

```json
{
  "port_cpu": 8004,
  "ip_cpu": "127.0.0.1",
  "ip_memory": "127.0.0.1",
  "port_memory": 8002,
  "ip_kernel": "127.0.0.1",
  "port_kernel": 8002,
  "tlb_entries": 15,
  "tlb_replacement": "LRU",
  "cache_entries": 10,
  "cache_replacement": "CLOCK",
  "cache_delay": 10,
  "log_level": "DEBUG"
}
```

---

## Módulo: Memoria + SWAP

Administra lectura/escritura en memoria y swap.

### Lineamiento e Implementación

* Servidor multihilo para Kernel y CPUs.
* Manejo de pseudocódigo y espacio usuario (`make([]byte, TamMemoria)`).

### Esquema de Memoria

* Paginación multinivel (configurable niveles y entradas).
* Swap en `swapfile.bin`.

#### Estructuras

* Array de bytes (usuario).
* Tablas de páginas (kernel).
* Archivo swap.
* Métricas por proceso.

#### Métricas

* Accesos a tablas de páginas.
* Instrucciones solicitadas.
* Bajadas a SWAP.
* Subidas a memoria principal.
* Lecturas/escrituras en memoria.

### Comunicación

* **Inicialización:** responde `OK` o error.
* **Suspensión:** liberar en memoria, escribir swap.
* **Des-suspensión:** leer swap, liberar swap, `OK`.
* **Finalización:** liberar espacios, log de métricas.
* **Acceso tabla:** retorna frame y cuenta niveles.
* **Acceso usuario:** `READ`/`WRITE` con `OK`.
* **Página completa:** `Leer` y `Actualizar` full page.
* **Memory Dump:** genera `<PID>-<TIMESTAMP>.dmp` en `dump_path`.
* **Swap:** manejar ubicaciones en swapfile.

### Logs mínimos y obligatorios

```text
## PID: <PID> - Proceso Creado - Tamaño: <TAM>
## PID: <PID> - Proceso Destruido - Métricas: ATP;<Inst>;<SWAP>;<MemPrin>;<LecMem>;<EscMem>
## PID: <PID> - Obtener instrucción: <PC> - Instrucción: <INSTR> <ARGS>
## PID: <PID> - <Escritura/Lectura> - Dir Física: <DIR> - Tamaño: <TAM>
## PID: <PID> - Memory Dump solicitado
```

### Archivo de Configuración

| Campo              | Tipo     | Descripción                   |
| :----------------- | :------- | :---------------------------- |
| port\_memory       | Numérico | Puerto de la Memoria          |
| ip\_memory         | String   | IP de la Memoria              |
| memory\_size       | Numérico | Bytes totales espacio usuario |
| page\_size         | Numérico | Tamaño de página              |
| entries\_per\_page | Numérico | Entradas por tabla            |
| number\_of\_levels | Numérico | Niveles de paginación         |
| memory\_delay      | Numérico | Delay respuesta memoria       |
| swapfile\_path     | String   | Path de `swapfile.bin`        |
| swap\_delay        | Numérico | Delay respuesta swap          |
| log\_level         | String   | Nivel de logs                 |
| dump\_path         | String   | Path para dumps               |
| scripts\_path      | String   | Path para scripts             |

### Ejemplo

```json
{
  "port_memory": 8002,
  "ip_memory": "127.0.0.1",
  "memory_size": 4096,
  "page_size": 64,
  "entries_per_page": 4,
  "number_of_levels": 5,
  "memory_delay": 1500,
  "swapfile_path": "/home/utnso/swapfile.bin",
  "swap_delay": 15000,
  "log_level": "DEBUG",
  "dump_path": "/home/utnso/dump_files/",
  "scripts_path": "/home/utnso/scripts/"
}
```