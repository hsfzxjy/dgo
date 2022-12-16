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
  String get pinTokenName => '_$entryName\$PinTokenImpl';
  String get constructorName => isEnum ? '.of' : '';
  String get renameTo => data['Rename'];

  OpStruct get struct => ir as OpStruct;
  Map<String, OpField> get structFields => struct.fields;

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
      vPort: '\$port',
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
      ..if_(!isEnum, _buildEqualAndHashCode)
      ..for_(methods, (method) => method.writeSnippet())
      ..if_(
          isPinnable,
          () => ctx
            ..sln('static \$dgo.PinToken<$entryName> Function(')
            ..sln('\$dgo.DgoPort, \$core.Iterator)')
            ..sln('\$dgoLoadPinToken = $pinTokenName.\$dgoLoad;'))
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
          ..then(_buildPinToken)
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
    ..sln('@\$meta.protected')
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
    ..sln('@\$meta.protected')
    ..sln('static ${ir.outerDartType} '
        '\$dgoLoad(\$dgo.DgoPort $vPort, \$core.Iterator<\$core.dynamic> $vArgs) {')
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
    ..sln('@\$meta.protected')
    ..sln('\$core.int '
        '\$dgoStore(\$core.List<\$core.dynamic> $vArgs, \$core.int $vIndex) {')
    ..sln('final $vHolder = this;')
    ..then(ir.writeSnippet$dgoStore)
    ..sln('return $vIndex;')
    ..sln('}');

  void _buildEqualAndHashCode() => ctx
    ..sln('@\$core.override')
    ..sln('\$core.bool operator==(\$core.Object other) {')
    ..sln('if (other is! $entryName) return false;')
    ..if_(
      ir is OpStruct,
      () => ctx
        ..for_(
          struct.fields.values,
          (f) => 'if (${f.name} != other.${f.name}) return false;',
        )
        ..sln('return true;'),
      else_: () => 'return \$inner == other.\$inner;',
    )
    ..sln('}')
    ..sln()
    ..sln('\$core.int get hashCode {')
    ..sln('\$core.int code = 0;')
    ..if_(
      ir is OpStruct,
      () =>
          ctx..for_(struct.fields.values, (f) => 'code ^= ${f.name}.hashCode;'),
      else_: () => 'code ^= \$inner.hashCode;',
    )
    ..sln('return code; }');

  void _buildPinToken() => ctx
    ..sln('class $pinTokenName extends \$dgo.PinToken<$entryName> {')
    ..sln('final \$dgo.DgoPort _port;')
    ..sln('final \$core.int _version;')
    ..sln('final \$core.int _lid;')
    ..sln('final \$core.int _key;')
    ..sln('final $entryName _data;')
    ..sln('\$core.bool _disposed = false;')
    ..sln()
    ..sln('$pinTokenName._(this._port, this._version, ')
    ..sln('this._lid, this._key, this._data);')
    ..sln()
    ..for_(
      struct.chans.values,
      (ch) => ctx
        ..sln('\$dgo.DartCallback? \$c${ch.chid}dcb;')
        ..sln('late final \$c${ch.chid} = ')
        ..sln('\$async.StreamController<${ch.dartType}>')
        ..sln(ch.isBroadcast ? '.broadcast' : '')
        ..sln('(onListen: _c${ch.chid}Listen, ')
        ..sln('onCancel: () {')
        ..sln('\$c${ch.chid}dcb?.remove();')
        ..sln('\$dgo.PreservedGoCall.chanCancelListen(_port, ')
        ..sln('[_version, _lid, _key, ${ch.chid}]);')
        ..sln('});')
        ..sln()
        ..sln('void _c${ch.chid}Listen() {')
        ..sln('void callback(\$core.Iterable args) {')
        ..sln('final $vArgs = args.iterator; $vArgs.moveNext();')
        ..sln('\$dgo.InvokeContext ctx = $vArgs.current; $vArgs.moveNext();')
        ..sln('if (ctx.flag.hasPop) { \$c${ch.chid}.close(); return; }')
        ..sln('${ch.dartType} $vHolder;')
        ..scope({}, ch.writeSnippet$dgoLoad)
        ..sln('\$c${ch.chid}.add($vHolder); }')
        ..sln('if (_disposed) { \$c${ch.chid}.close(); return; }')
        ..sln('\$c${ch.chid}dcb = _port.pend(callback);')
        ..sln('\$dgo.PreservedGoCall.chanListen(_port, ')
        ..sln('[_version, _lid, _key, ${ch.chid}, \$c${ch.chid}dcb!.id]); }'),
    )
    ..sln()
    ..sln('@\$core.override $entryName dispose() {')
    ..sln('if (_disposed) {')
    ..sln("throw 'dgo:dart: PinToken.dispose() '")
    ..sln("'must be called for exactly once'; }")
    ..sln('_disposed = true;')
    ..for_(struct.chans.values,
        (ch) => ctx..sln('\$c${ch.chid}dcb?.remove(); \$c${ch.chid}.close();'))
    ..sln('\$dgo.PreservedGoCall.tokenDispose(_port, [_version, _lid, _key]);')
    ..sln('return _data; }')
    ..sln()
    ..sln(
        '@\$core.override \$core.int \$dgoStore(\$core.List args, \$core.int index) {')
    ..sln('args[index] = _version;')
    ..sln('args[index + 1] = _lid;')
    ..sln('args[index + 2] = _key;')
    ..sln('return index + 3; }')
    ..sln()
    ..sln('@\$core.override \$core.String toString() => ')
    ..sln("'\$runtimeType(\$_version, \$_lid, \$_key)';")
    ..sln()
    ..sln('static \$dgo.PinToken<$entryName> \$dgoLoad(')
    ..sln('\$dgo.DgoPort port, \$core.Iterator args) {')
    ..sln('final version = args.current;')
    ..sln('args.moveNext();')
    ..sln('final lid = args.current;')
    ..sln('args.moveNext();')
    ..sln('final key = args.current;')
    ..sln('args.moveNext();')
    ..sln('final data = $entryName.\$dgoLoad(port, args);')
    ..sln('return $pinTokenName._(port, version, lid, key, data); }')
    ..sln('}');

  void _buildPinTokenExtension() {
    final ir = this.ir as OpStruct;
    ctx
      ..sln('extension $entryName\$PinTokenExt on \$dgo.PinToken<$entryName> {')
      ..sln('$pinTokenName get _token {')
      ..sln('final token = this as $pinTokenName;')
      ..sln('if (token._disposed) {')
      ..sln("throw 'dgo:dart: \$token is disposed'; }")
      ..sln('return token; }')
      ..sln()
      ..for_(
        struct.chans.values,
        (ch) => ctx
          ..sln('\$async.Stream<${ch.dartType}>')
          ..sln('get ${ch.name} => _token.\$c${ch.chid}.stream;'),
      )
      ..sln()
      ..for_(
        ir.fields.values,
        (field) => ctx
          ..sln('${field.dartType} get ${field.name} =>')
          ..sln(' _token._data.${field.name};')
          ..sln(),
      )
      ..for_(
        methods,
        (method) => ctx..then(method.writePinTokenSnippet),
      )
      ..sln('}');
  }
}
