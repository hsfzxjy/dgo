// Dart imports:
import 'dart:ffi';

// Package imports:
import 'package:dgo/dgo.dart';
import 'package:test/test.dart';

// Project imports:
import 'package:test_gen_dart/generated/test_gen_go/module.dart';

final dlib = DynamicLibrary.open('../_build/libtest_gen.so');

void main() async {
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
  });

  group('TesterWithField', () {
    test('ReturnsSelf', () async {
      final result = await TesterWithField(42).ReturnsSelf();
      expect(result.field, 42);
    });
  });
}
