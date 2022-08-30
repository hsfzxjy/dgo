library dgo;

// Dart imports:
import 'dart:ffi';
import 'dart:isolate';

// Package imports:
import 'package:meta/meta.dart';

// Project imports:
import './dgo_binding.dart' as binding;

part 'dylib.dart';
part 'callback_flag.dart';
part 'dart_callback.dart';
part 'go_callback.dart';

class Dgo {
  static void init(DynamicLibrary dylib) => _init(dylib);
  static int pendDart(Function fn) => _dartCallbackPend(fn);
  static void shutdown() => _receivePort.close();

  @visibleForTesting
  static bool dartCallbackExist(int id) {
    return _dartCallbackMap.containsKey(id);
  }
}
