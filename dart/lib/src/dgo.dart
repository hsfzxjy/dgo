library dgo;

// Dart imports:
import 'dart:async';
import 'dart:ffi';
import 'dart:isolate';
import 'dart:typed_data';

// Package imports:
import 'package:logging/logging.dart';
import 'package:meta/meta.dart';

// Project imports:
import './dgo_binding.dart' as binding;

part 'callback_flag.dart';
part 'dart_callback.dart';
part 'go_callback.dart';
part 'go_object.dart';
part 'port.dart';
part 'special_int.dart';
part 'invoke_context.dart';
part 'utils.dart';
part 'math_utils.dart';
part 'pin.dart';
part 'preserved_go_call.dart';

class Dgo extends DgoPortLike with _PortMixin {
  DgoPort? _defaultPort;
  DgoPort get defaultPort => _defaultPort!;
  @override
  DgoPort get _port => defaultPort;

  Future<DgoPort> initDefaultPort(DynamicLibrary dylib) =>
      DgoPort._build('dgo:dart:default', dylib, isDefault: true);

  Future<DgoPort> newPort(DynamicLibrary dylib, {String name = 'custom'}) =>
      DgoPort._build('dgo:dart:$name', dylib, isDefault: false);
}

final dgo = Dgo();
