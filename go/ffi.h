#ifndef DGO_VALUE_H
#define DGO_VALUE_H

#include "dart_api.h"
#include "dart_api_dl.h"
#include "dart_native_api.h"

typedef struct _dgo__Dart_CObject {
  Dart_CObject_Type Type;
  union {
    bool    as_bool;
    int32_t as_int32;
    int64_t as_int64;
    double  as_double;
    char *  as_string;
    struct {
      Dart_Port id;
      Dart_Port origin_id;
    } as_send_port;
    struct {
      int64_t id;
    } as_capability;
    struct {
      intptr_t                    length;
      struct _dgo__Dart_CObject **values;
    } as_array;
    struct {
      Dart_TypedData_Type type;
      intptr_t            length; /* in elements, not bytes */
      uint8_t *           values;
    } as_typed_data;
    struct {
      Dart_TypedData_Type  type;
      intptr_t             length; /* in elements, not bytes */
      uint8_t *            data;
      void *               peer;
      Dart_HandleFinalizer callback;
    } as_external_typed_data;
    struct {
      intptr_t             ptr;
      intptr_t             size;
      Dart_HandleFinalizer callback;
    } as_native_pointer;
  } Value;
} dgo__Dart_CObject;

typedef struct {
  Dart_TypedData_Type Type;
  intptr_t            Length;
  uint8_t *           Values;
} dgo__Dart_CObject_AsTypedData;

typedef struct {
  intptr_t       Length;
  Dart_CObject **Values;
} dgo__Dart_CObject_AsArray;

typedef struct {
  Dart_TypedData_Type Type;
  intptr_t            Length; /* in elements, not bytes */
  uint8_t *           Data;
  uintptr_t           Peer;     // mod: from void*
  uintptr_t           Callback; // mod: from Dart_HandleFinalizer
} dgo__Dart_Cobject_AsExternalTypedData;

void         dgo__InitFFI(void *data);
Dart_Port_DL dgo__InitPort(Dart_Port_DL send_port_id);
bool dgo__PostCObjects(Dart_Port_DL port_id, int cnt, dgo__Dart_CObject *cobjs);
bool dgo__PostCObjects2(
    Dart_Port_DL       port_id,
    int                cnt1,
    dgo__Dart_CObject *cobjs1,
    int                cnt2,
    dgo__Dart_CObject *cobjs2);
bool dgo__PostCObject(Dart_Port_DL port_id, dgo__Dart_CObject *cobj);
bool dgo__PostInt(Dart_Port_DL port_id, int64_t value);
bool dgo__CloseNativePort(Dart_Port_DL port_id);

extern uintptr_t dgo__pGoFinalizer;

#endif