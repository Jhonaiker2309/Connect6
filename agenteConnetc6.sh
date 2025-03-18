#!/usr/bin/env bash

# Script: agenteConnetc6.sh
# Uso:    ./agenteConnetc6.sh -fichas=negras -tpj=4

# Valores por defecto:
fichas="negras"
tiempo=4

# Parseo sencillo de argumentos:
for ARG in "$@"; do
  case $ARG in
    -fichas=*)
      fichas="${ARG#*=}"
      ;;
    -tpj=*)
      tiempo="${ARG#*=}"
      ;;
    *)
      echo "Error: argumento desconocido: $ARG"
      exit 1
      ;;
  esac
done

# Llamamos al binario con los flags deseados:
# Asumiendo que tu ejecutable final está en bin/connect6
# y que éste acepta flags --fichas y --tpj

exec ./bin/connect6 -fichas="$fichas" -tpj="$tiempo"
