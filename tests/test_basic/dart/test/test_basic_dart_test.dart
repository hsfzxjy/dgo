// Dart imports:
import 'dart:async';
import 'dart:ffi' as ffi;

// Package imports:
import 'package:dgo/dgo.dart';
import 'package:dgo/src/dgo.dart' show DgoPortState;
import 'package:logging/logging.dart';
import 'package:quiver/collection.dart';
import 'package:test/test.dart' as t;
import 'package:test/test.dart' show group;

// Project imports:
import 'package:test_basic_dart/binding.dart';

part 'barrier.dart';

extension<T> on Future<T> {
  Future<T> defaultTimeout() => timeout(Duration(seconds: 1));
}

final dlib = ffi.DynamicLibrary.open('../_build/libtest_basic.so');
final lib = LibTestBasic(dlib);

void setupLogger() {
  Logger.root.level = Level.ALL;
  Logger.root.onRecord.listen((record) {
    print(
      '${record.level.name}: ${record.time}: ${record.message}\n'
      'error=${record.error}\n'
      '${record.stackTrace}',
    );
  });
}

void main() async {
  setupLogger();
  var port = await dgo.initDefaultPort(dlib);
  testWithPort(port, 'Default Port');
  port = await dgo.newPort(dlib);
  testWithPort(port, 'Custom Port');
}

void testWithPort(DgoPort port, String testPrefix) {
  final completers = <int, Completer>{};

  Barrier.blockHere(setup: () {
    lib.InitTestContext(
        port.pend((int token) {
          completers[token]!.complete();
        }).id,
        port.key,
        port.isDefault);
  });

  group('[$testPrefix] Test Correctness:', () {
    Future<void> runCases(Function nativeFunction, Function echoer) async {
      var exhausted = false;
      while (!exhausted) {
        final com = Completer();
        final token = port.pend(echoer).id;
        completers[token] = com;
        if (1 == nativeFunction(token)) {
          exhausted = true;
        }
        await com.future.defaultTimeout();
      }
    }

    test('Single Value', () async {
      await runCases(lib.TestSingle, (GoCallback gcb, dynamic x) {
        gcb.flag(CF.pop).call([x]);
      });
    });

    test('Tuple Value', () async {
      await runCases(lib.TestTuple, (GoCallback gcb, dynamic x1, dynamic x2) {
        gcb.flag(CF.pop).call([x1, x2]);
      });
    });
  });

  group('[$testPrefix] Test Dart Flag:', () {
    test('Pop', () async {
      final com = Completer();
      var counter = 0;
      late DartCallback dcb;
      dcb = port.pend(() {
        counter++;
        if (counter == 2) {
          assert(!dcb.exists);
          com.complete();
        }
      });
      lib.TestDartPop(dcb.id);
      return com.future.defaultTimeout();
    });

    test('PackArray', () async {
      final com = Completer();
      late DartCallback dcb;
      dcb = port.pend((List args) {
        assert(!dcb.exists);
        assert(listsEqual(args, [1, 'hello', 3.14]));
        com.complete();
      });
      lib.TestDartPackArray(dcb.id);
      return com.future.defaultTimeout();
    });

    test('WithContext', () async {
      final com = Completer();
      late DartCallback dcb;
      dcb = port.pend((InvokeContext context, int a1, String a2, double a3) {
        final cf = context.flag;
        assert(!dcb.exists);
        assert(cf.hasPop && cf.hasWithContext);
        assert(a1 == 1);
        assert(a2 == 'hello');
        assert(a3 == 3.14);
        com.complete();
      });
      lib.TestDartWithContext(dcb.id);
      return com.future.defaultTimeout();
    });

    test('Fast', () async {
      final com = Completer();
      int counter = 0;
      final answers = [null, false, true];

      late DartCallback dcb;
      dcb = port.pend((dynamic a) {
        assert(a == answers[counter]);
        counter++;
        if (counter == 3) {
          assert(!dcb.exists);
          com.complete();
        }
      });
      lib.TestDartFast(dcb.id);
      return com.future.defaultTimeout();
    });

    test('FastVoid', () async {
      final com = Completer();
      late DartCallback dcb;
      dcb = port.pend(() {
        assert(!dcb.exists);
        com.complete();
      });
      lib.TestDartFastVoid(dcb.id);
      return com.future.defaultTimeout();
    });

    test('Fallible', () async {
      final dcb = port.pend(() {});
      dcb.remove();
      lib.TestDartFallible(dcb.id);
      return Future.delayed(Duration(milliseconds: 10));
    });
  });

  group('[$testPrefix] Test Dart Future:', () {
    test('Resolve', () async {
      final com = Completer();
      final dcb = port.pendFuture(com);
      lib.TestDartFutureResolve(dcb.id);
      final x = await com.future;
      assert(x == 42);
      assert(!dcb.exists);
    });

    test('Reject', () async {
      final com = Completer();
      final dcb = port.pendFuture(com);
      lib.TestDartFutureReject(dcb.id);
      try {
        await com.future;
      } catch (e) {
        assert(e == 'this is an error');
      }
      assert(!dcb.exists);
    });
  });

  group('[$testPrefix] Test Dart Stream:', () {
    test('Stream', () async {
      final con = StreamController();
      final dcb = port.pendStream(con);
      lib.TestDartStream(dcb.id);
      final com = Completer();
      var counter = 0;
      con.stream.listen(
        (value) {
          switch (counter) {
            case 0:
              assert(value == 1);
              break;
            case 1:
              assert(value == 3.14);
              break;
            case 3:
              assert(value == '4', value);
              break;
            default:
              throw 'counter = $counter, value = $value';
          }
          counter++;
        },
        onError: (e) {
          assert(counter == 2);
          assert(e == 'error 1');
          counter++;
        },
        onDone: () => com.complete(),
      );
      await com.future;
      assert(!dcb.exists);
    });
  });

  group('[$testPrefix] Test Go Flag:', () {
    test('Pop', () async {
      final com = Completer();
      final dcb = port.pend((GoCallback gcb) {
        gcb.flag(CF).call();
        gcb.flag(CF.pop).call();
      });
      completers[dcb.id] = com;
      lib.TestGoPop(dcb.id);
      return com.future.defaultTimeout();
    });

    test('WithContext', () async {
      final com = Completer();
      final dcb = port.pend((GoCallback gcb) {
        gcb.flag(CF.pop.withContext).call([1, 'hello', 3.14, null]);
      });
      completers[dcb.id] = com;
      lib.TestGoWithContext(dcb.id);
      return com.future.defaultTimeout();
    });

    test('PackArray', () async {
      final com = Completer();
      final dcb = port.pend((GoCallback gcb) {
        gcb.flag(CF.pop.withContext.packArray).call([1, 'hello', 3.14]);
      });
      completers[dcb.id] = com;
      lib.TestGoPackArray(dcb.id);
      return com.future.defaultTimeout();
    });

    test('Fast', () async {
      final com = Completer();
      final dcb = port.pend((GoCallback gcb) {
        gcb.flag(CF.fast(CFFK.nil)).call();
        gcb.flag(CF.fast(CFFK.no)).call();
        gcb.flag(CF.pop.fast(CFFK.yes)).call();
      });
      completers[dcb.id] = com;
      lib.TestGoFast(dcb.id);
      return com.future.defaultTimeout();
    });

    test('FastVoid', () async {
      final com = Completer();
      final dcb = port.pend((GoCallback gcb) {
        gcb.flag(CF.fast(CFFK.void_).pop).call();
      });
      completers[dcb.id] = com;
      lib.TestGoFastVoid(dcb.id);
      return com.future.defaultTimeout();
    });

    test('Fallible', () async {
      final com = Completer();
      final dcb = port.pend((GoCallback gcb) {
        gcb.flag(CF.pop.fallible).call();
        com.complete();
      });
      lib.TestGoFallible(dcb.id);
      return com.future.defaultTimeout();
    });
  });

  test('PortClosed', () async {
    final isDefault = port.isDefault ? 1 : 0;
    final key = port.key;
    await port.close();
    assert(port.state == DgoPortState.closed);
    lib.TestPortClosed(key, isDefault);
    GoCallback.fromRaw(0, dgoPort: port).flag(CF).call();
  });
}
