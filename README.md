# Sistema Operativo Distribuido en Go

![Go Version](https://img.shields.io/badge/Go-1.24.0-blue.svg)
![License](https://img.shields.io/badge/License-Academic-green.svg)

## Descripción

Este proyecto implementa un **sistema operativo distribuido** desarrollado en Go como parte del trabajo práctico de la materia Sistemas Operativos. El sistema simula los componentes principales de un SO moderno con una arquitectura distribuida que permite ejecutar los módulos en diferentes máquinas virtuales.

### Arquitectura

El sistema está compuesto por cuatro módulos principales que se comunican mediante APIs HTTP:

```
    ┌─────────┐
    │ KERNEL  │◄─────────────┐
    └────┬────┘              │
         │                   │
    ┌────▼────┐         ┌────┴────┐
    │   CPU   │         │   I/O   │
    └────┬────┘         └─────────┘
         │
    ┌────▼────┐
    │ MEMORIA │
    └─────────┘
```

## Módulos del Sistema

### Kernel
- **Función**: Gestión de procesos y planificación
- **Características**:
  - Diagrama de 7 estados (NEW, READY, EXEC, BLOCKED, EXIT, SUSP.READY, SUSP.BLOCKED)
  - Planificadores de corto, mediano y largo plazo
  - Algoritmos: FIFO, SJF, SRT, PMCP
  - PCB con métricas de estado y tiempo

### CPU
- **Función**: Ejecución de instrucciones
- **Características**:
  - Soporte para múltiples instancias
  - TLB (Translation Lookaside Buffer) con algoritmos FIFO/LRU
  - Cache de memoria con algoritmos CLOCK/CLOCK-M
  - Manejo de interrupciones

### Memoria
- **Función**: Gestión de memoria virtual y física
- **Características**:
  - Paginación multinivel
  - Sistema de SWAP
  - Gestión de marcos y páginas
  - Métricas de rendimiento

### I/O
- **Función**: Simulación de dispositivos de entrada/salida
- **Características**:
  - Dispositivos configurables (DISCO, etc.)
  - Operaciones asíncronas
  - Múltiples instancias simultáneas

## 📁 Estructura del Proyecto

```
├── bin/                    # Ejecutables compilados
├── cmd/                    # Código fuente de módulos
│   ├── cpu/               # Módulo CPU
│   ├── io/                # Módulo I/O
│   ├── kernel/            # Módulo Kernel
│   └── memoria/           # Módulo Memoria
├── configs/               # Archivos de configuración
├── scripts/               # Scripts de pseudocódigo
├── swap/                  # Archivos de intercambio
└── utils/                 # Utilidades compartidas
```


### Compilación

```bash
# Clonar el repositorio
git clone --depth 1 --branch main https://github.com/GonzaloPontnau/SISTEMA-OPERATIVO.go.git.git
cd tp-2025-1c-LosCuervosXeneizes

# Compilar todos los módulos
go build -o ./bin/memoria ./cmd/memoria
go build -o ./bin/kernel ./cmd/kernel
go build -o ./bin/cpu ./cmd/cpu
go build -o ./bin/io ./cmd/io
```

## Uso

### Configuración de Red Distribuida

Para ejecutar en múltiples máquinas, actualizar las IPs en los archivos de configuración:

```bash
# Cambiar IPs según la configuración de red
sed -i 's/"IP_CPU": *"127\.0\.0\.1"/"IP_CPU": "192.168.0.127"/g' configs/*.json
sed -i 's/"IP_MEMORIA": *"127\.0\.0\.1"/"IP_MEMORIA": "192.168.0.190"/g' configs/*.json
sed -i 's/"IP_KERNEL": *"127\.0\.0\.1"/"IP_KERNEL": "192.168.0.107"/g' configs/*.json
sed -i 's/"IP_IO": *"127\.0\.0\.1"/"IP_IO": "192.168.0.107"/g' configs/*.json
```

### Ejecución Básica

El orden de inicio es importante debido a las dependencias:

1. **Memoria** (debe iniciarse primero)
```bash
./bin/memoria configs/memoria-config-EstabilidadGeneral.json
```

2. **I/O** (dispositivos necesarios)
```bash
./bin/io DISCO configs/io1-config-EstabilidadGeneral.json
```

3. **CPU** (una o más instancias)
```bash
./bin/cpu CPU1 configs/cpu1-config-EstabilidadGeneral.json
./bin/cpu CPU2 configs/cpu2-config-EstabilidadGeneral.json
```

4. **Kernel** (con proceso inicial)
```bash
./bin/kernel configs/kernel-config-EstabilidadGeneral.json scripts/ESTABILIDAD_GENERAL 0
```

## Pruebas del Sistema

### Prueba de Planificación de Corto Plazo

Evalúa algoritmos FIFO, SJF y SRT con múltiples CPUs:

```bash
# Terminal 1: Memoria
./bin/memoria configs/memoria-config-PlaniCorto.json

# Terminal 2: I/O
./bin/io DISCO1 configs/io1-config-PlaniCorto.json

# Terminal 3-4: CPUs
./bin/cpu CPU1 configs/cpu1-config-PlaniCorto.json
./bin/cpu CPU2 configs/cpu2-config-PlaniCorto.json

# Terminal 5: Kernel con FIFO
./bin/kernel configs/kernel-config-PlaniCortoFIFO.json scripts/PLANI_CORTO_PLAZO 0
```

### Prueba de Planificación de Mediano/Largo Plazo

Evalúa algoritmos de admisión FIFO y PMCP:

```bash
# Configuración similar pero con configs específicos
./bin/kernel configs/kernel-config-PlaniMedianoLargoFIFO.json scripts/PLANI_LYM_PLAZO 0
```

### Prueba de Memoria y SWAP

Evalúa el sistema de memoria virtual:

```bash
./bin/memoria configs/memoria-config-MemoriaSWAP.json
./bin/kernel configs/kernel-config-MemoriaSWAP.json scripts/MEMORIA_IO 90
```

### Prueba de TLB

Evalúa algoritmos de reemplazo de TLB (FIFO/LRU):

```bash
./bin/cpu CPU1 configs/cpu1-config-TLB.json  # FIFO
./bin/cpu CPU1 configs/cpu2-config-TLB.json  # LRU
```

### Prueba de Cache

Evalúa algoritmos de cache (CLOCK/CLOCK-M):

```bash
./bin/cpu CPU1 configs/cpu1-config-MemoriaCache.json  # CLOCK
./bin/cpu CPU1 configs/cpu2-config-MemoriaCache.json  # CLOCK-M
```

## Configuración

### Parámetros de Kernel
- `ALGORITMO_CORTO_PLAZO`: FIFO, SJF, SRT
- `ALGORITMO_INGRESO_A_READY`: FIFO, PMCP
- `GRADO_MULTIPROGRAMACION`: Número máximo de procesos en memoria
- `ALFA`: Factor de suavizado para SJF/SRT
- `ESTIMACION_INICIAL`: Estimación inicial para algoritmos predictivos

### Parámetros de CPU
- `ENTRADAS_TLB`: Número de entradas en TLB
- `REEMPLAZO_TLB`: FIFO o LRU
- `ENTRADAS_CACHE`: Número de entradas en cache
- `REEMPLAZO_CACHE`: CLOCK o CLOCK-M

### Parámetros de Memoria
- `TAM_MEMORIA`: Tamaño total de memoria física
- `TAM_PAGINA`: Tamaño de cada página
- `ENTRADAS_POR_TABLA`: Entradas por tabla de páginas
- `CANTIDAD_NIVELES`: Niveles de paginación

## Logging y Métricas

El sistema genera logs detallados con nivel configurable:
- **INFO**: Información general del sistema
- **DEBUG**: Información detallada para debugging
- **ERROR**: Errores del sistema

### Métricas de PCB
Al finalizar cada proceso se muestran:
- Conteo de transiciones entre estados
- Tiempo total en cada estado
- Tiempo de espera promedio


## Troubleshooting

### Problemas Comunes

1. **Error de conexión**: Verificar que las IPs y puertos sean correctos
2. **Módulo no responde**: Verificar el orden de inicio (Memoria → I/O → CPU → Kernel)
3. **Falta de memoria**: Ajustar `GRADO_MULTIPROGRAMACION` o `TAM_MEMORIA`

### Logs de Depuración

Cambiar `LOG_LEVEL` a `DEBUG` en los archivos de configuración para obtener información detallada.

> [!NOTE]
> Este proyecto es desarrollado con fines académicos como parte del trabajo práctico de Sistemas Operativos para la UTN.

---
