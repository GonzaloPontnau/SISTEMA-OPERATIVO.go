#!/bin/bash

# Script para iniciar Prueba de Estabilidad General con logs
echo "=== Iniciando Prueba de Estabilidad General ==="

# Crear directorio de logs
mkdir -p logs

# Limpiar logs anteriores (opcional)
echo "Limpiando logs anteriores..."
rm -f logs/*.log

echo "Iniciando módulos..."

# Iniciar Memoria PRIMERO
echo "▶ Iniciando Memoria..."
./bin/memoria configs/memoria-config.json > logs/memoria.log 2>&1 &
MEMORIA_PID=$!
sleep 3

# Iniciar IOs SEGUNDO
echo "▶ Iniciando IO1 (DISCO1)..."
./bin/io DISCO1 configs/io1-config.json > logs/io1.log 2>&1 &
IO1_PID=$!

echo "▶ Iniciando IO2 (DISCO2)..."
./bin/io DISCO2 configs/io2-config.json > logs/io2.log 2>&1 &
IO2_PID=$!

echo "▶ Iniciando IO3 (DISCO3)..."
./bin/io DISCO3 configs/io3-config.json > logs/io3.log 2>&1 &
IO3_PID=$!

echo "▶ Iniciando IO4 (DISCO4)..."
./bin/io DISCO4 configs/io4-config.json > logs/io4.log 2>&1 &
IO4_PID=$!

sleep 2

# Iniciar CPUs TERCERO
echo "▶ Iniciando CPU1..."
./bin/cpu CPU1 configs/cpu1-config.json > logs/cpu1.log 2>&1 &
CPU1_PID=$!

echo "▶ Iniciando CPU2..."
./bin/cpu CPU2 configs/cpu2-config.json > logs/cpu2.log 2>&1 &
CPU2_PID=$!

echo "▶ Iniciando CPU3..."
./bin/cpu CPU3 configs/cpu3-config.json > logs/cpu3.log 2>&1 &
CPU3_PID=$!

echo "▶ Iniciando CPU4..."
./bin/cpu CPU4 configs/cpu4-config.json > logs/cpu4.log 2>&1 &
CPU4_PID=$!

sleep 2

# Iniciar Kernel ÚLTIMO
echo "▶ Iniciando Kernel..."
./bin/kernel scripts/ESTABILIDAD_GENERAL 0 > logs/kernel.log 2>&1 &
KERNEL_PID=$!

sleep 3

echo ""
echo "✅ Todos los módulos iniciados!"
echo ""
echo "📁 Logs guardándose en:"
echo "   - logs/memoria.log"
echo "   - logs/kernel.log" 
echo "   - logs/cpu1.log, cpu2.log, cpu3.log, cpu4.log"
echo "   - logs/io1.log, io2.log, io3.log, io4.log"
echo ""
echo "📊 Para monitorear en tiempo real:"
echo "   tail -f logs/kernel.log | grep -E '(ERROR|CPU[1-4]|DISCO[1-4])'"
echo ""
echo "💾 Para monitorear recursos:"
echo "   watch -n 30 \"ps aux | grep -E 'bin/(memoria|kernel|cpu|io)' | grep -v grep\""
echo ""
echo "🛑 Para detener todo:"
echo "   ./stop_test.sh"
echo ""

# Guardar PIDs para poder detener después
echo "MEMORIA_PID=$MEMORIA_PID" > .test_pids
echo "KERNEL_PID=$KERNEL_PID" >> .test_pids
echo "CPU1_PID=$CPU1_PID" >> .test_pids
echo "CPU2_PID=$CPU2_PID" >> .test_pids
echo "CPU3_PID=$CPU3_PID" >> .test_pids
echo "CPU4_PID=$CPU4_PID" >> .test_pids
echo "IO1_PID=$IO1_PID" >> .test_pids
echo "IO2_PID=$IO2_PID" >> .test_pids
echo "IO3_PID=$IO3_PID" >> .test_pids
echo "IO4_PID=$IO4_PID" >> .test_pids

echo "⏱️  Prueba en curso... Monitorear por 10-15 minutos."