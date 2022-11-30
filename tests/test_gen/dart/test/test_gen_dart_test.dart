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

    test('ReturnsExternalType', () async {
      final result = await Tester().ReturnsExternalType();
      expect(result.FieldString, 'string');
    });
  });
}
