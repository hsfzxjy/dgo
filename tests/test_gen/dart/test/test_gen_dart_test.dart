@Timeout(Duration(seconds: 5))

// Dart imports:
import 'dart:async';
import 'dart:ffi';

// Package imports:
import 'package:dgo/dgo.dart';
import 'package:test/test.dart';

// Project imports:
import 'package:test_gen_dart/generated/registrar.dart';
import 'package:test_gen_dart/generated/test_gen_go/module.dart';

import 'package:test_gen_dart/generated/test_gen_go/internal/module.dart'
    show StructWithSimpleTypes;

final dlib = DynamicLibrary.open('../_build/libtest_gen.so');

void main() async {
  registerDgoRelated();
  await dgo.initDefaultPort(dlib);

  group('Tester', () {
    test('ReturnsVoid', () async {
      await Tester().ReturnsVoid();
      expectAsync0(() {})();
    });

    test('ReturnsString', () async {
      expect(await Tester().ReturnsString(), 'Hello world!');
    });

    test('ReturnsError', () async {
      expectLater(() => Tester().ReturnsError(), throwsA(equals('error')));
    });

    test('ReturnsStringOrError', () async {
      expect(await Tester().ReturnsStringOrError(true), 'success');
      expectLater(
          () => Tester().ReturnsStringOrError(false), throwsA(equals('error')));
    });

    test('ReturnsSlice', () async {
      final f = Tester().ReturnsSlice;
      expect(await f(1), equals(['0']));
      expect(await f(2), equals(['0', '1']));
    });

    test('ReturnsIdentitySlice', () async {
      final f = Tester().ReturnsIdentitySlice;
      expect(await f([42, 84]), equals([42, 84]));
    });

    test('ReturnsMap', () async {
      final f = Tester().ReturnsMap;
      expect(await f(1), equals({0: '0'}));
      expect(await f(2), equals({0: '0', 1: '1'}));
    });

    test('ReturnsIdentityMap', () async {
      final f = Tester().ReturnsIdentityMap;
      expect(await f({42: '84'}), equals({42: '84'}));
    });

    test('ReturnsExternalType', () async {
      final result = await Tester().ReturnsExternalType();
      expect(result.FieldString, 'string');
    });

    test('ReturnsStructWithSimpleTypes', () async {
      final obj = StructWithSimpleTypes(-12, -24, -36, -48, -60, 12, 24, 36, 48,
          60, 72, 3, 3.14, true, 'string');
      final obj2 = await Tester().ReturnsStructWithSimpleTypes(obj);
      expect(obj, equals(obj2));
    });
  });

  group('TesterWithField', () {
    test('ReturnsSelf', () async {
      final result = await TesterWithField(42).ReturnsSelf();
      expect(result.field, 42);
    });
  });

  group('PinTester', () {
    void testP(String description, Function(PinToken<Peripheral>) body) {
      test(description, () async {
        final gc = Completer<bool>();
        final p =
            await PinTester().MakeAndReturnsPeripheral(dgo.pendFuture(gc).id);
        await body(p);
        for (final i in Iterable.generate(5)) {
          final timeLimit = Duration(milliseconds: 100 << i);
          await PinTester().GC();
          final gcSuccess =
              await gc.future.timeout(timeLimit, onTimeout: () => false);
          if (gcSuccess) return;
          print('p is not GCed within $timeLimit');
        }
        await expectLater(gc.future, completes);
      });
    }

    testP('token invalid after GC', (p) async {
      expect(p.id, 42);
      expect(p.name, 'MyDevice');
      final result = await PinTester().AcceptPeripheralAndCompute(p);
      final expected = 'Peripheral<id=42, name=MyDevice>';
      expect(result, expected);
      expect(await p.ToString(), expected);
      p.dispose();
      expect(
        p.dispose,
        throwsA(matches(RegExp(r'^dgo:dart:.*exactly once$'))),
      );
      await PinTester().GC();
      await PinTester().AssertTokenInvalid(p);
    });

    testP('chan functionality', (p) async {
      final stream = p.state;
      final fut = expectLater(stream, emitsInOrder([1, 2, 3, emitsDone]));
      await PinTester().StartStateAndUnpin(p, true, true);
      await fut;
      p.dispose();
      await PinTester().AssertTokenInvalid(p);
    });

    testP('stream done even chan not closed', (p) async {
      final stream = p.state.asBroadcastStream();
      final fut = expectLater(stream, emitsInOrder([1, 2, 3]));
      await PinTester().StartStateAndUnpin(p, true, false);
      await fut;
      p.dispose();
      await expectLater(stream, emitsDone);
      await PinTester().AssertTokenInvalid(p);
    });

    testP('stream directly done after chan closed', (p) async {
      final stream = p.state.asBroadcastStream();
      await PinTester().StartStateAndUnpin(p, false, true);
      await expectLater(stream, emitsDone);
      p.dispose();
      await expectLater(stream, emitsDone);
      await PinTester().AssertTokenInvalid(p);
    });

    testP('stream done after token disposed', (p) async {
      final state = p.state;
      await PinTester().StartStateAndUnpin(p, true, false);
      p.dispose();
      final fut = expectLater(state, emitsDone);
      await fut;
    });

    testP('blockUntilListen', (p) async {
      final state = p.stateBlock;
      await PinTester().StartStateBlockAndUnpin(p);
      await expectLater(state, emitsInOrder([1, 2, 3, emitsDone]));
      p.dispose();
    });

    testP('broadcast', (p) async {
      final state = p.stateBroadcast;
      final fut = expectLater(state, emitsInOrder([1, 2, 3, emitsDone]));
      await PinTester().StartStateBroadcastAndUnpin(p);
      await fut;
      expect(state.isBroadcast, isTrue);
      p.dispose();
    });

    testP('memorized', (p) async {
      final state = p.stateMemo;
      await PinTester().StartStateMemoAndUnpin(p);
      await expectLater(state, emitsInOrder([3]));
      p.dispose();
    });
  });
}
