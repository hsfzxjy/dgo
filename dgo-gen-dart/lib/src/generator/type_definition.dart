part of 'generator.dart';

class EnumMember {
  final String name;
  final String value;

  EnumMember(JsonMap m)
      : name = m['Name'],
        value = m['Value'];
}

class TypeDefinition {
  final FileSet fileset;
  final JsonMap data;

  final int typeId;
  final Namable ir;

  final bool isEnum;
  final bool isPinnable;

  EntryUri get myUri => ir.myUri!;
  File get file => fileset[myUri.goMod.dartModFile];
  String get entryName => myUri.name;
  String get constructorName => isEnum ? '.of' : '';
  String get renameTo => data['Rename'];

  Map<String, OpField> get structFields => (ir as OpStruct).fields;

  Iterable<Method> get methods => (data['Methods'] as List)
      .cast<JsonMap>()
      .map((m) => Method.fromMap(ir, m));

  Iterable<EnumMember> get enumMembers =>
      (data['EnumMembers'] as List).cast<JsonMap>().map(EnumMember.new);

  TypeDefinition(this.fileset, this.data)
      : ir = IR.fromJSON(data['Term']) as Namable,
        typeId = data['TypeId'],
        isEnum = data['IsEnum'],
        isPinnable = data['IsPinnable'] {
    assert(ir.isNamed);
  }

  Future<void> save() async {
    setFile(file, {
      vArgs: '\$args',
      vIndex: '\$index',
      vHolder: '\$o',
      vSize: '\$size',
    });

    ctx
      ..if_(
        isEnum,
        _buildHeaderEnum,
        else_: _buildHeaderClass,
      )
      ..sln('static const typeId = $typeId;')
      ..if_(
        ir is OpStruct,
        _buildConstructorStruct,
        else_: _buildConstructorOther,
      )
      ..scope({}, _build$dgoLoad)
      ..scope({}, _build$dgoStore)
      ..scope({}, _build$dgoGoSize)
      ..for_(methods, (method) => method.writeSnippet())
      ..sln('}')
      ..if_(
        renameTo.isNotEmpty,
        () => ctx
          ..sln()
          ..sln('typedef $renameTo = $entryName;'),
      )
      ..if_(
        isPinnable,
        () => ctx
          ..sln()
          ..then(_buildPinTokenExtension),
      );
  }

  void _buildHeaderEnum() => ctx
    ..sln('enum $entryName implements \$dgo.DgoObject {')
    ..sln(enumMembers.map((m) => '${m.name}(${m.value})').joinComma)
    ..sln(';')
    ..sln('factory $entryName.of(\$core.int value) {')
    ..sln('switch (value) {')
    ..for_(
      enumMembers,
      (m) => ctx
        ..sln('case ${m.value}:')
        ..sln('return ${m.name};'),
    )
    ..sln('default:')
    ..sln("throw 'dgo:dart: cannot convert \$value to $entryName';")
    ..sln('}')
    ..sln('}')
    ..sln();

  void _buildHeaderClass() => ctx
    ..sln('@\$meta.immutable')
    ..if_(
      isPinnable,
      () => ctx.sln('class $entryName extends \$dgo.Pinnable {'),
      else_: () => ctx..sln('class $entryName extends \$dgo.DgoObject {'),
    );

  void _buildConstructorStruct() => ctx
    ..for_(
      (ir as OpStruct).fields.values,
      (field) => ctx.sln('final ${field.term.dartType} ${field.name};'),
    )
    ..str('const $entryName(')
    ..sln(structFields.values.map((field) => 'this.${field.name}').joinComma)
    ..sln(');');

  void _buildConstructorOther() => ctx
    ..sln('final ${ir.dartType} \$inner;')
    ..sln('const $entryName(this.\$inner);');

  void _build$dgoGoSize() => ctx
    ..sln('@\$core.override')
    ..if_(
      ir.isGoNotDynamic,
      () => ctx
        ..sln('final \$dgoGoSize = ${ir.goSize};')
        ..sln(),
      else_: () => ctx
        ..sln('\$core.int get \$dgoGoSize ')
        ..sln('{\$core.int $vSize = 0;')
        ..sln('final $vHolder = this;')
        ..then(ir.writeSnippet$dgoGoSize)
        ..sln('return $vSize; }'),
    );

  void _build$dgoLoad() => ctx
    ..sln()
    ..sln('static ${ir.outerDartType} '
        '\$dgoLoad(\$core.Iterator<\$core.dynamic> $vArgs) {')
    ..sln('${ir.dartType} $vHolder;')
    ..scope({}, ir.writeSnippet$dgoLoad)
    ..if_(
      ir is OpStruct,
      () => ctx.sln('return $vHolder;'),
      else_: () => ctx.sln('return $entryName$constructorName($vHolder);'),
    )
    ..sln('}');

  void _build$dgoStore() => ctx
    ..sln('@\$core.override')
    ..sln('\$core.int '
        '\$dgoStore(\$core.List<\$core.dynamic> $vArgs, \$core.int $vIndex) {')
    ..sln('final $vHolder = this;')
    ..then(ir.writeSnippet$dgoStore)
    ..sln('return $vIndex;')
    ..sln('}');

  void _buildPinTokenExtension() {
    final ir = this.ir as OpStruct;
    ctx
      ..sln('extension $entryName\$PinTokenExt on \$dgo.PinToken<$entryName> {')
      ..for_(
        ir.fields.values,
        (field) => ctx
          ..sln('${field.dartType} get ${field.name} =>')
          ..sln(' \$dgo.checkTokenValidity(this, \$data.${field.name});')
          ..sln(),
      )
      ..for_(
        methods,
        (method) => ctx
          ..sln(method.signature('Function'))
          ..sln('get ${method.funcName} =>')
          ..sln('\$dgo.checkTokenValidity(this, \$data.${method.funcName});')
          ..sln(),
      )
      ..sln('}');
  }
}
