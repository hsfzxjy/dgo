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
}

extension _IRExt on IR {
  void writeSnippetLoad(GeneratorContext ctx) {
    if (this is Namable && (this as Namable).isNamed) {
      final myUri = (this as Namable).myUri!;
      ctx.buffer
        ..writeln('{')
        ..writeln('${ctx[vHolder]} = ')
        ..writeln(ctx.importer.qualifyUri(myUri))
        ..writeln('.\$dgoLoad(${ctx[vArgs]});')
        ..writeln('}');
    } else {
      writeSnippet$dgoLoad(ctx);
    }
  }
}

extension _SumExt<T extends num> on Iterable<T> {
  T sum() => isEmpty ? 0 as T : reduce((a, b) => (a + b) as T);
}

class Method {
  final IR self;
  final String funcName;
  final List<Param> params;
  final int funcId;
  final IR? returnType;
  final bool returnError;

  Method.fromMap(this.self, Map m)
      : funcName = m['Name'],
        funcId = m['FuncId'],
        params = (m['Params'] as List).map((m) => Param.fromMap(m)).toList(),
        returnType = _buildIRNull(m['Return']),
        returnError = m['ReturnError'];

  void writeSnippet(GeneratorContext ctx) {
    var paramSig = params
        .map((p) => '${p.term.outerDartType(ctx.importer)} ${p.name}')
        .followedBy(['{Duration? \$timeout, DgoPort? \$port}']).join(',');
    var paramSize = params.map((p) => p.term.goSize).sum();
    paramSize += self.goSize;
    ctx.buffer
      ..writeln('Future<${returnType.dartType(ctx.importer)}>')
      ..writeln('$funcName($paramSig) async {')
      ..writeln('\$port ??= dgo.defaultPort;')
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
      ..writeln('final \$future = GoMethod($funcId, \$port)'
          '.${returnType == null ? "call" : "callWithResult"}'
          '(${ctx[vArgs]}, '
          'timeout: \$timeout, hasError: $returnError);')
      ..if_(
          returnType == null,
          () => ctx.buffer.writeln('return \$future;'),
          () => ctx.buffer
            ..writeln('{')
            ..writeln('final ${ctx[vArgs]} = (await \$future).iterator;')
            ..writeln('${ctx[vArgs]}.moveNext();')
            ..writeln('${returnType!.dartType(ctx.importer)} ${ctx[vHolder]};')
            ..pipe(returnType!.writeSnippetLoad(ctx))
            ..writeln('return ${ctx[vHolder]};')
            ..writeln('}'))
      ..writeln('}');
  }
}
