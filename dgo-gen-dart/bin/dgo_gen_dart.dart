// Dart imports:
import 'dart:convert';
import 'dart:io' as io;

// Project imports:
import 'package:dgo_gen_dart/dgo_gen_dart.dart';

void main(List<String> arguments) async {
  final irFile = io.File('../tests/gen_tests_dart/ir.json');
  final ir = jsonDecode(await irFile.readAsString());
  final genDir = io.Directory('../tests/gen_tests_dart/lib/generated');
  try {
    await genDir.delete(recursive: true);
  } on io.FileSystemException catch (e) {
    if (e.osError?.errorCode != 2) {
      rethrow;
    }
  }
  await genDir.create(recursive: true);
  Generator(ir, genDir.path).save();
}
