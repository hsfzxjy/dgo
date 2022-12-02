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

class Generator {
  final JsonMap definitions;
  final FileSet fileSet = FileSet();
  Generator(JsonMap payload) : definitions = payload['Definitions'] {
    config = Config(payload['Config'])..validate();
  }

  void _processEntry(JsonMap entry) {
    final typeId = entry['TypeId'] as int;
    final ir = IR.fromJSON(entry['Term']);
    assert(ir is Namable && ir.myUri != null);
    final myUri = (ir as Namable).myUri!;
    final file = fileSet[myUri.goMod.dartModFile];
    final entryName = myUri.name;

    final isEnum = entry['IsEnum'] as bool;
    final enumMembers = entry['EnumMembers'] as List;
    final constructorName = isEnum ? '.of' : '';

    setFile(file, {
      vArgs: '\$args',
      vIndex: '\$startIndex',
      vHolder: '\$instance',
    });

    ctx
      ..if_(
        isEnum,
        () => ctx
          ..sln('enum $entryName {')
          ..sln(
              enumMembers.map((m) => "${m['Name']}(${m['Value']})").join(', '))
          ..sln(';')
          ..sln('factory $entryName.of(\$core.int value) {')
          ..sln('switch (value) {')
          ..for_(enumMembers, (m) {
            ctx
              ..sln("case ${m['Value']}:")
              ..sln("return ${m['Name']};");
          })
          ..sln('default:')
          ..sln("throw 'dgo:dart: cannot convert \$value to $entryName';")
          ..sln('}')
          ..sln('}')
          ..sln(),
        else_: () => ctx
          ..sln('@immutable')
          ..sln('${isEnum ? "enum" : "class"} $entryName {'),
      )
      ..sln('static const typeId = $typeId;');

    ctx.if_(
      ir is OpStruct,
      () => ctx
        ..for_(
          (ir as OpStruct).fields.values,
          (field) => ctx.sln('final ${field.term.dartType} ${field.name};'),
        )
        ..str('const $entryName(')
        ..sln(ir.fields.values.map((field) => 'this.${field.name}').join(','))
        ..sln(');'),
      else_: () => ctx
        ..sln('final ${ir.dartType} \$inner;')
        ..sln('const $entryName(this.\$inner);'),
    );

    ctx
      ..sln()
      ..sln(
          'static ${ir.outerDartType} \$dgoLoad(\$core.Iterator<\$core.dynamic> $vArgs){')
      ..sln('${ir.dartType} $vHolder;')
      ..scope({}, ir.writeSnippet$dgoLoad)
      ..if_(
        ir is OpStruct,
        () => ctx.sln('return $vHolder;'),
        else_: () => ctx.sln('return $entryName$constructorName($vHolder);'),
      )
      ..sln('}');

    for (final methodSpec in entry['Methods'] ?? []) {
      Method.fromMap(ir, methodSpec).writeSnippet(ctx);
    }

    ctx
      ..sln()
      ..sln(
          '\$core.int \$dgoStore(\$core.List<\$core.dynamic> $vArgs, \$core.int $vIndex) {')
      ..sln('final $vHolder = this;')
      ..scope({}, ir.writeSnippet$dgoStore)
      ..sln('return $vIndex;')
      ..sln('}')
      ..sln('}');

    final rename = entry['Rename'] as String;
    if (rename.isNotEmpty) {
      ctx
        ..sln()
        ..sln('typedef $rename = $entryName;');
    }
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
    for (final entry in definitions.values) {
      _processEntry(entry);
    }
    await fileSet.save();
  }
}
