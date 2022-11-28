// Dart imports:
import 'dart:convert';
import 'dart:io' as io;

// Project imports:
import 'package:dgo_gen_dart/dgo_gen_dart.dart';

void main(List<String> arguments) async {
  final rawPayload = await io.stdin.transform(utf8.decoder).join('');
  final payload = jsonDecode(rawPayload);
  await Generator(payload).save();
}
