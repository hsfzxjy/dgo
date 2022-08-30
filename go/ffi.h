#ifndef DGO_VALUE_H
#define DGO_VALUE_H

#include "dart_api.h"
#include "dart_api_dl.h"
#include "dart_native_api.h"

typedef struct {
  Dart_TypedData_Type type;
  intptr_t            length;
  uint8_t *           values;
} dgo__Dart_CObject_AsTypedData;

typedef struct {
  intptr_t       length;
  Dart_CObject **values;
} dgo__Dart_CObject_AsArray;

typedef struct {
  Dart_TypedData_Type type;
  intptr_t            length; /* in elements, not bytes */
  uint8_t *           data;
  uintptr_t           peer;     // mod: from void*
  uintptr_t           callback; // mod: from Dart_HandleFinalizer
} dgo__Dart_Cobject_AsExternalTypedData;

Dart_Port_DL dgo_InitFFI(void *data, Dart_Port_DL sendPort);
bool         dgo__PostCObjects(int cnt, Dart_CObject *cobjs);
bool         dgo__PostInt(int64_t);

extern uintptr_t dgo__pGoFinalizer;

#endif