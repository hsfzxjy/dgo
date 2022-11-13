part of 'ir.dart';

class Param {
  final String name;
  final IR term;

  Param.fromMap(Map m)
      : term = _buildIR(m['Term']),
        name = m['Name'];
}

extension _NullableIRExt on IR? {
  String dartType(Importer context) =>
      this == null ? 'void' : this!.dartType(context);
  String holderType(Importer context) =>
      this == null ? 'void' : 'List<dynamic>';
}

extension _IRExt on IR {
  void writeSnippetLoad(GeneratorContext ctx) {
    if (this is Namable && (this as Namable).isNamed) {
      final myUri = (this as Namable).myUri!;
      ctx.buffer
        ..writeln('{')
        ..writeln('final \$result = ')
        ..writeln(ctx.importer.qualifyUri(myUri))
        ..writeln('.\$dgoLoad(${ctx[vArgs]},${ctx[vIndex]});')
        ..writeln('${ctx[vHolder]} = \$result.result;')
        ..writeln('${ctx[vIndex]} = \$result.nextIndex;')
        ..writeln('}');
    } else {
      writeSnippet$dgoLoad(ctx);
    }
  }
}

class Method {
  final IR self;
  final String funcName;
  final List<Param> params;
  final int funcId;
  final IR? returnType;

  Method.fromMap(this.self, Map m)
      : funcName = m['Name'],
        funcId = m['FuncId'],
        params = (m['Params'] as List).map((m) => Param.fromMap(m)).toList(),
        returnType = _buildIRNull(m['Return']);

  void writeSnippet(GeneratorContext ctx) {
    var paramSig = params
        .map((p) => '${p.term.dartType(ctx.importer)} ${p.name}')
        .join(',');
    var paramSize = params.map((p) => p.term.goSize).reduce((a, b) => a + b);
    paramSize += self.goSize;
    ctx.buffer
      ..writeln('Future<${returnType.dartType(ctx.importer)}>')
      ..writeln('$funcName($paramSig, {Duration? \$timeout}) async {')
      ..writeln(
          'final ${ctx[vArgs]} = List<dynamic>.filled($paramSize, null, growable: false);')
      ..writeln('var ${ctx[vIndex]} = 0;')
      ..writeln('${ctx[vIndex]} = \$dgoStore(${ctx[vArgs]}, ${ctx[vIndex]});')
      ..for_(
          params,
          (param) => ctx.buffer
            ..writeln('{')
            ..writeln('final ${ctx[vHolder]} = ${param.name};')
            ..pipe(param.term.writeSnippet$dgoStore(ctx))
            ..writeln('}'))
      ..writeln(
          'final \$completer = Completer<${returnType.holderType(ctx.importer)}>();')
      ..writeln('final \$callback = Dgo.pendCompleter(\$completer);')
      ..writeln(
          'Future<${returnType.holderType(ctx.importer)}> \$future = \$completer.future;')
      ..writeln('if (\$timeout != null) {')
      ..writeln('\$future = \$future.timeout(\$timeout, onTimeout: () async {')
      ..writeln('Dgo.removeDart(\$callback);')
      ..writeln("throw 'The Go call fails to respond in \${\$timeout}';")
      ..writeln('});')
      ..writeln('}')
      ..writeln(
          'GoMethod($funcId).call(${ctx[vArgs]}, ${ctx[vIndex]}, \$callback);')
      ..if_(
          returnType == null,
          () => ctx.buffer.writeln('return \$completer.future;'),
          () => ctx.buffer
            ..writeln('{')
            ..writeln('final ${ctx[vArgs]} = await \$completer.future;')
            ..writeln('var ${ctx[vIndex]} = 0;')
            ..writeln('${returnType!.dartType(ctx.importer)} ${ctx[vHolder]};')
            ..pipe(returnType!.writeSnippetLoad(ctx))
            ..writeln('return ${ctx[vHolder]};')
            ..writeln('}'))
      ..writeln('}');
  }
}
