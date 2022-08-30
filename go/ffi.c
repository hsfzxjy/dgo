#include "ffi.h"

#include <stdlib.h>

bool Dgo_CloseNativePort_DL(Dart_Port_DL port) {
  return Dart_CloseNativePort_DL(port);
}

static Dart_Port_DL dartSendPort = ILLEGAL_PORT;
static Dart_Port_DL dartReceivePort = ILLEGAL_PORT;

extern void dgo__HandleNativeMessage(Dart_Port_DL, Dart_CObject *);
extern void dgo__GoFinalizer(uintptr_t, uintptr_t);

uintptr_t dgo__pGoFinalizer = (uintptr_t)(&dgo__GoFinalizer);

Dart_Port_DL dgo_InitFFI(void *data, Dart_Port_DL sendPort) {
  if (dartReceivePort != ILLEGAL_PORT) {
    Dart_CloseNativePort_DL(sendPort);
  }
  dartSendPort = sendPort;
  Dart_InitializeApiDL(data);
  dartReceivePort =
      Dart_NewNativePort_DL("dgo_port", dgo__HandleNativeMessage, false);
  Dart_CObject arg;
  arg.type = Dart_CObject_kSendPort;
  arg.value.as_send_port.id = dartReceivePort;
  Dart_PostCObject_DL(sendPort, &arg);
}

const int DEFAULT_ARGS_SIZE = 10;

bool dgo__PostInt(int64_t obj) {
  if (dartSendPort == ILLEGAL_PORT)
    return false;

  return Dart_PostInteger_DL(dartSendPort, obj);
}

bool dgo__PostCObjects(int cnt, Dart_CObject *cobjs) {
  if (dartSendPort == ILLEGAL_PORT)
    return false;

  Dart_CObject   arg;
  Dart_CObject * pargs[DEFAULT_ARGS_SIZE];
  Dart_CObject **ppargs, **allocated = NULL;
  if (cnt <= DEFAULT_ARGS_SIZE) {
    ppargs = &pargs[0];
    allocated = false;
  } else {
    allocated = ppargs = calloc(cnt, sizeof(Dart_CObject *));
  }
  arg.type = Dart_CObject_kArray;
  arg.value.as_array.length = cnt;
  arg.value.as_array.values = ppargs;

  for (int i = 0; i < cnt; i++) {
    *ppargs = &cobjs[0];
    ppargs++;
    cobjs++;
  }

  Dart_PostCObject_DL(dartSendPort, &arg);

  if (allocated != NULL)
    free(allocated);

  return true;
}

#include "dart_api_dl.c"