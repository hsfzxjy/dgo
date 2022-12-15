#include "ffi.h"

#include <stdlib.h>

extern void dgo__HandleNativeMessage(Dart_Port_DL, Dart_CObject *);
extern void dgo__GoFinalizer(uintptr_t, uintptr_t);

uintptr_t dgo__pGoFinalizer = (uintptr_t)(&dgo__GoFinalizer);

void dgo__InitFFI(void *data) { Dart_InitializeApiDL(data); }

Dart_Port_DL dgo__InitPort(Dart_Port_DL send_port_id) {
  Dart_Port_DL receive_port_id =
      Dart_NewNativePort_DL("dgo:go:", dgo__HandleNativeMessage, false);
  Dart_CObject arg;
  arg.type = Dart_CObject_kSendPort;
  arg.value.as_send_port.id = receive_port_id;
  Dart_PostCObject_DL(send_port_id, &arg);
  return receive_port_id;
}

bool dgo__PostCObject(Dart_Port_DL port_id, dgo__Dart_CObject *cobj) {
  return Dart_PostCObject_DL(port_id, (Dart_CObject *)cobj);
}

bool dgo__PostInt(Dart_Port_DL port_id, int64_t value) {
  return Dart_PostInteger_DL(port_id, value);
}

bool dgo__PostCObjects(
    Dart_Port_DL       port_id,
    int                cnt,
    dgo__Dart_CObject *cobjs) {
  return dgo__PostCObjects2(port_id, cnt, cobjs, 0, NULL);
}

bool dgo__PostCObjects2(
    Dart_Port_DL       port_id,
    int                cnt1,
    dgo__Dart_CObject *cobjs1,
    int                cnt2,
    dgo__Dart_CObject *cobjs2) {

  const int DEFAULT_ARGS_SIZE = 10;

  dgo__Dart_CObject   arg;
  dgo__Dart_CObject * pargs[DEFAULT_ARGS_SIZE];
  dgo__Dart_CObject **ppargs, **allocated = NULL;

  int cnt = cnt1 + cnt2;
  if (cnt <= DEFAULT_ARGS_SIZE) {
    ppargs = &pargs[0];
    allocated = false;
  } else {
    allocated = ppargs = calloc(cnt, sizeof(dgo__Dart_CObject *));
  }
  arg.Type = Dart_CObject_kArray;
  arg.Value.as_array.length = cnt;
  arg.Value.as_array.values = ppargs;

  for (int i = 0; i < cnt1; i++) {
    *ppargs = &cobjs1[0];
    ppargs++;
    cobjs1++;
  }

  for (int i = cnt1; i < cnt; i++) {
    *ppargs = &cobjs2[0];
    ppargs++;
    cobjs2++;
  }

  Dart_PostCObject_DL(port_id, (Dart_CObject *)&arg);

  if (allocated != NULL)
    free(allocated);

  return true;
}

bool dgo__CloseNativePort(Dart_Port_DL port_id) {
  return Dart_CloseNativePort_DL(port_id);
}

#include "dart_api_dl.c"