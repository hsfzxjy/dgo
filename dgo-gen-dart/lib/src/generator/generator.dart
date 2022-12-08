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

  void _saveRegistrar() {
    final file = fileSet[GoModUri.empty.dartFile('registrar.dart')];

    setFile(file, {});

    final typeNames =
        definitions.keys.map(EntryUri.fromString).map(ctx.importer.qualifyUri);

    ctx
      ..sln('\$dgo.DgoObject _buildObjectById(')
      ..sln('\$core.int typeId, \$core.Iterator args) {')
      ..sln('switch (typeId) {')
      ..for_(
        typeNames,
        (typeName) => ctx
          ..sln('case $typeName.typeId:')
          ..sln('return $typeName.\$dgoLoad(args);'),
      )
      ..sln('default:')
      ..sln("throw 'dgo:dart: cannot build object for typeId=\$typeId'; } }");

    ctx
      ..sln('T _buildObject<T extends \$dgo.DgoObject>(')
      ..sln('\$core.Iterator args) {')
      ..sln('switch (T) {')
      ..for_(
        typeNames,
        (typeName) => ctx
          ..sln('case $typeName:')
          ..sln('return $typeName.\$dgoLoad(args) as T;'),
      )
      ..sln('default:')
      ..sln("throw 'dgo:dart: cannot build object for type=\$T'; } }");

    ctx
      ..sln('void registerDgoRelated() {')
      ..sln('\$dgo.registerTypes(')
      ..sln('_buildObjectById, _buildObject); }');
  }

  Future<void> save() async {
    await _prepareDirectory();
    for (final definition in definitions.values) {
      TypeDefinition(fileSet, definition).save();
    }
    _saveRegistrar();
    await fileSet.save();
  }
}
