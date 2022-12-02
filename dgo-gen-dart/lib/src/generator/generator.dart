// Dart imports:
import 'dart:io' as io;

// Package imports:
import 'package:dart_style/dart_style.dart';
import 'package:meta/meta.dart';
import 'package:path/path.dart' as p;

// Project imports:
import '../ir/ir.dart';
import '../misc.dart';
import '../uri.dart';

part 'importer.dart';
part 'context.dart';
part 'context_ext.dart';
part 'file.dart';
part 'fileset.dart';
part 'config.dart';
part 'type_definition.dart';

class Generator {
  final JsonMap definitions;
  final FileSet fileSet = FileSet();
  Generator(JsonMap payload) : definitions = payload['Definitions'] {
    config = Config(payload['Config'])..validate();
  }

  Future<void> _prepareDirectory() async {
    final directory = io.Directory(config.generatedInPath);
    try {
      await directory.delete(recursive: true);
    } on io.FileSystemException catch (e) {
      if (e.osError?.errorCode != 2) {
        rethrow;
      }
    }
    await directory.create(recursive: true);
  }

  Future<void> save() async {
    await _prepareDirectory();
    for (final definition in definitions.values) {
      TypeDefinition(fileSet, definition).save();
    }
    await fileSet.save();
  }
}
