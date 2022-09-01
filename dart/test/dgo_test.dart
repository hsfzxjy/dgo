// Dart imports:
import 'dart:async';
import 'dart:ffi' as ffi;

// Package imports:
import 'package:quiver/collection.dart';
import 'package:test/test.dart';

// Project imports:
import 'package:dgo/dgo.dart';

void Function(int) lookupVoidU32Func(ffi.DynamicLibrary dylib, String name) =>
    dylib
        .lookup<ffi.NativeFunction<ffi.Void Function(ffi.Uint32)>>(name)
        .asFunction<void Function(int)>();

int Function(int) lookupIntU32Func(ffi.DynamicLibrary dylib, String name) =>
    dylib
        .lookup<ffi.NativeFunction<ffi.Int Function(ffi.Uint32)>>(name)
        .asFunction<int Function(int)>();

Future<void> timeout(Future fut) {
  return Future.any([
    Future.delayed(Duration(seconds: 1)).then((_) => Future.error('timeout')),
    fut,
  ]);
}

void main() {
  final dylib = ffi.DynamicLibrary.open('../build/libtest.so');
  Dgo.init(dylib);

  final setResolveCallback = lookupVoidU32Func(dylib, 'SetResolveCallback');

  final comMap = <int, Completer>{};

  setResolveCallback(Dgo.pendDart((int token) {
    comMap[token]!.complete();
  }));

  Future<void> runCases(String nativeMethod, Function echoer) {
    var exhausted = false;

    Future<void> testOnce() {
      final com = Completer();
      final token = Dgo.pendDart(echoer);
      comMap[token] = com;
      if (1 == lookupIntU32Func(dylib, nativeMethod)(token)) {
        exhausted = true;
      }

      return timeout(com.future);
    }

    final futs = <Future>[];
    while (!exhausted) {
      futs.add(testOnce());
    }

    return Future.wait(futs);
  }

  group('Test Correctness:', () {
    test('Single Value', () async {
      await runCases('TestSingle', (int gcb, dynamic x) {
        GoCallback(gcb).flag(CallbackFlag().pop()).call([x]);
      });
    });

    test('Tuple Value', () async {
      await runCases('TestTuple', (int gcb, dynamic x1, dynamic x2) {
        GoCallback(gcb).flag(CallbackFlag().pop()).call([x1, x2]);
      });
    });
  });

  group('Test Dart Flag:', () {
    test('Pop', () async {
      final com = Completer();
      var counter = 0;
      int dcb = 0;
      dcb = Dgo.pendDart(() {
        counter++;
        if (counter == 2) {
          assert(!Dgo.dartCallbackExist(dcb));
          com.complete();
        }
      });
      lookupIntU32Func(dylib, 'TestDartPop')(dcb);
      return timeout(com.future);
    });

    test('PackArray', () async {
      final com = Completer();
      int dcb = 0;
      dcb = Dgo.pendDart((List args) {
        assert(!Dgo.dartCallbackExist(dcb));
        assert(listsEqual(args, [1, 'hello', 3.14]));
        com.complete();
      });
      lookupIntU32Func(dylib, 'TestDartPackArray')(dcb);
      return timeout(com.future);
    });

    test('WithCode', () async {
      final com = Completer();
      int dcb = 0;
      dcb = Dgo.pendDart((CallbackFlag cf, int a1, String a2, double a3) {
        assert(!Dgo.dartCallbackExist(dcb));
        assert(cf.hasPop && cf.hasWithCode);
        assert(a1 == 1);
        assert(a2 == 'hello');
        assert(a3 == 3.14);
        com.complete();
      });
      lookupIntU32Func(dylib, 'TestDartWithCode')(dcb);
      return timeout(com.future);
    });

    test('Fast', () async {
      final com = Completer();
      int counter = 0;
      final answers = [null, false, true];

      int dcb = 0;
      dcb = Dgo.pendDart((dynamic a) {
        assert(a == answers[counter]);
        counter++;
        if (counter == 3) {
          assert(!Dgo.dartCallbackExist(dcb));
          com.complete();
        }
      });
      lookupIntU32Func(dylib, 'TestDartFast')(dcb);
      return timeout(com.future);
    });

    test('FastVoid', () async {
      final com = Completer();
      int dcb = 0;
      dcb = Dgo.pendDart(() {
        assert(!Dgo.dartCallbackExist(dcb));
        com.complete();
      });
      lookupIntU32Func(dylib, 'TestDartFastVoid')(dcb);
      return timeout(com.future);
    });
  });

  group('Test Dart Future:', () {
    test('Resolve', () async {
      final com = Completer();
      final dcb = Dgo.pendCompleter(com);
      lookupIntU32Func(dylib, 'TestDartFutureResolve')(dcb);
      final x = await com.future;
      assert(x == 42);
      assert(!Dgo.dartCallbackExist(dcb));
    });

    test('Reject', () async {
      final com = Completer();
      final dcb = Dgo.pendCompleter(com);
      lookupIntU32Func(dylib, 'TestDartFutureReject')(dcb);
      try {
        await com.future;
      } catch (e) {
        assert(e == 'this is an error');
      }
      assert(!Dgo.dartCallbackExist(dcb));
    });
  });

  group('Test Go Flag:', () {
    test('Pop', () async {
      final com = Completer();
      final dcb = Dgo.pendDart((int gcb) {
        GoCallback(gcb).flag(CF).call([]);
        GoCallback(gcb).flag(CF.pop()).call([]);
      });
      comMap[dcb] = com;
      lookupIntU32Func(dylib, 'TestGoPop')(dcb);
      return timeout(com.future);
    });

    test('WithCode', () async {
      final com = Completer();
      final dcb = Dgo.pendDart((int gcb) {
        GoCallback(gcb)
            .flag(CF.pop().withCode())
            .call([1, 'hello', 3.14, null]);
      });
      comMap[dcb] = com;
      lookupIntU32Func(dylib, 'TestGoWithCode')(dcb);
      return timeout(com.future);
    });

    test('PackArray', () async {
      final com = Completer();
      final dcb = Dgo.pendDart((int gcb) {
        GoCallback(gcb)
            .flag(CF.pop().withCode().packArray())
            .call([1, 'hello', 3.14]);
      });
      comMap[dcb] = com;
      lookupIntU32Func(dylib, 'TestGoPackArray')(dcb);
      return timeout(com.future);
    });

    test('Fast', () async {
      final com = Completer();
      final dcb = Dgo.pendDart((int gcb) {
        GoCallback(gcb).flag(CF.fast(CFFK.nil)).call([]);
        GoCallback(gcb).flag(CF.fast(CFFK.no)).call([]);
        GoCallback(gcb).flag(CF.fast(CFFK.yes).pop()).call([]);
      });
      comMap[dcb] = com;
      lookupIntU32Func(dylib, 'TestGoFast')(dcb);
      return timeout(com.future);
    });

    test('FastVoid', () async {
      final com = Completer();
      final dcb = Dgo.pendDart((int gcb) {
        GoCallback(gcb).flag(CF.fast(CFFK.void_).pop()).call([]);
      });
      comMap[dcb] = com;
      lookupIntU32Func(dylib, 'TestGoFastVoid')(dcb);
      return timeout(com.future);
    });
  });
}
