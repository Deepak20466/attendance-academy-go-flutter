import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/widgets/async_value_widget.dart';
import '../../classes/data/class_repository.dart';
import '../data/student_repository.dart';

class StudentsScreen extends ConsumerWidget {
  const StudentsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final students = ref.watch(studentsListProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Students')),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showCreateDialog(context, ref),
        icon: const Icon(Icons.add),
        label: const Text('Add student'),
      ),
      body: RefreshIndicator(
        onRefresh: () => ref.refresh(studentsListProvider.future),
        child: AsyncValueWidget(
          value: students,
          data: (result) {
            if (result.data.isEmpty) {
              return ListView(
                children: const [
                  Padding(
                    padding: EdgeInsets.all(32),
                    child: Center(child: Text('No students found.', style: TextStyle(color: Colors.grey))),
                  ),
                ],
              );
            }
            return ListView.separated(
              padding: const EdgeInsets.symmetric(vertical: 8),
              itemCount: result.data.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (context, index) {
                final student = result.data[index];
                return ListTile(
                  leading: CircleAvatar(child: Text(student.name.isNotEmpty ? student.name[0].toUpperCase() : '?')),
                  title: Text(student.name),
                  subtitle: Text(student.guardianPhone.isNotEmpty ? student.guardianPhone : student.phone),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () => context.push('/students/${student.id}'),
                );
              },
            );
          },
        ),
      ),
    );
  }

  Future<void> _showCreateDialog(BuildContext context, WidgetRef ref) async {
    final batches = await ref.read(classRepositoryProvider).listBatches();
    if (!context.mounted) return;

    if (batches.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Create a batch first before adding students.')),
      );
      return;
    }

    final nameController = TextEditingController();
    final phoneController = TextEditingController();
    final guardianNameController = TextEditingController();
    final guardianPhoneController = TextEditingController();
    String? selectedBatchId = batches.first.id;
    String? errorText;

    await showDialog<void>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => AlertDialog(
          title: const Text('Add student'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                TextField(controller: nameController, decoration: const InputDecoration(labelText: 'Name')),
                const SizedBox(height: 12),
                DropdownButtonFormField<String>(
                  initialValue: selectedBatchId,
                  decoration: const InputDecoration(labelText: 'Batch'),
                  items: batches.map((b) => DropdownMenuItem(value: b.id, child: Text(b.name))).toList(),
                  onChanged: (v) => setState(() => selectedBatchId = v),
                ),
                const SizedBox(height: 12),
                TextField(controller: phoneController, decoration: const InputDecoration(labelText: 'Phone')),
                const SizedBox(height: 12),
                TextField(controller: guardianNameController, decoration: const InputDecoration(labelText: 'Guardian name')),
                const SizedBox(height: 12),
                TextField(controller: guardianPhoneController, decoration: const InputDecoration(labelText: 'Guardian phone')),
                if (errorText != null) ...[
                  const SizedBox(height: 12),
                  Text(errorText!, style: const TextStyle(color: Colors.red)),
                ],
              ],
            ),
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(context), child: const Text('Cancel')),
            FilledButton(
              onPressed: () async {
                if (nameController.text.trim().isEmpty || selectedBatchId == null) {
                  setState(() => errorText = 'Name and batch are required.');
                  return;
                }
                try {
                  await ref.read(studentRepositoryProvider).create(
                        batchId: selectedBatchId!,
                        name: nameController.text.trim(),
                        phone: phoneController.text.trim(),
                        guardianName: guardianNameController.text.trim(),
                        guardianPhone: guardianPhoneController.text.trim(),
                      );
                  ref.invalidate(studentsListProvider);
                  if (context.mounted) Navigator.pop(context);
                } catch (e) {
                  setState(() => errorText = 'Failed to create student: $e');
                }
              },
              child: const Text('Add'),
            ),
          ],
        ),
      ),
    );
  }
}
