part of dgo;

late final binding.LibDgo _lib;

late final ReceivePort _receivePort;
SendPort? _sendPort;

bool _initialized = false;

void _init(DynamicLibrary dylib) {
  if (_initialized) throw 'dgo:dart already initialized';
  _initialized = true;

  _receivePort = ReceivePort('dgo:dart');
  _receivePort.listen(_handleMessage);

  _lib = binding.LibDgo(dylib);
  _lib.dgo_InitFFI(
      NativeApi.initializeApiDLData, _receivePort.sendPort.nativePort);
  _lib.dgo_InitGo();
}

void _handleMessage(dynamic msg) {
  if (msg is SendPort) {
    if (_sendPort != null) {
      throw 'dgo:dart received multiple SendPort';
    }
    _sendPort = msg;
  } else {
    if (_sendPort == null) throw 'dgo:dart sendPort not initalized';
    if (msg is List) {
      _dartCallbackHandle(msg);
    } else if (msg is int) {
      _dartCallbackHandle([msg]);
    } else {
      throw 'dgo:dart unsupported message $msg';
    }
  }
}
