import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../analytics/data/analytics_repository.dart';
import '../../auth/data/auth_controller.dart';
import '../../classes/data/class_repository.dart';
import '../../classes/presentation/classes_screen.dart';
import '../../coaches/data/coach_repository.dart';
import '../../students/presentation/students_screen.dart';
import '../../coaches/presentation/coaches_screen.dart';
import 'admin_dashboard_screen.dart';
import 'coach_dashboard_screen.dart';
import 'more_menu_screen.dart';

/// Root authenticated shell: a bottom-nav scaffold whose tabs differ by
/// role, since a coach has no business seeing a coach roster or
/// cross-activity management screens.
class HomeShell extends ConsumerStatefulWidget {
  const HomeShell({super.key});

  @override
  ConsumerState<HomeShell> createState() => _HomeShellState();
}

class _HomeShellState extends ConsumerState<HomeShell> {
  int _index = 0;

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(authControllerProvider).valueOrNull;
    if (session == null) return const SizedBox.shrink();

    final isAdmin = session.isAdmin;

    final pages = isAdmin
        ? const [
            AdminDashboardScreen(),
            StudentsScreen(),
            CoachesScreen(),
            ClassesScreen(),
            MoreMenuScreen(),
          ]
        : const [
            CoachDashboardScreen(),
            ClassesScreen(),
            StudentsScreen(),
            MoreMenuScreen(),
          ];

    final destinations = isAdmin
        ? const [
            NavigationDestination(icon: Icon(Icons.dashboard_outlined), selectedIcon: Icon(Icons.dashboard), label: 'Dashboard'),
            NavigationDestination(icon: Icon(Icons.groups_outlined), selectedIcon: Icon(Icons.groups), label: 'Students'),
            NavigationDestination(icon: Icon(Icons.badge_outlined), selectedIcon: Icon(Icons.badge), label: 'Coaches'),
            NavigationDestination(icon: Icon(Icons.event_note_outlined), selectedIcon: Icon(Icons.event_note), label: 'Classes'),
            NavigationDestination(icon: Icon(Icons.more_horiz), label: 'More'),
          ]
        : const [
            NavigationDestination(icon: Icon(Icons.dashboard_outlined), selectedIcon: Icon(Icons.dashboard), label: 'Dashboard'),
            NavigationDestination(icon: Icon(Icons.event_note_outlined), selectedIcon: Icon(Icons.event_note), label: 'My Classes'),
            NavigationDestination(icon: Icon(Icons.groups_outlined), selectedIcon: Icon(Icons.groups), label: 'Students'),
            NavigationDestination(icon: Icon(Icons.more_horiz), label: 'More'),
          ];

    final safeIndex = _index >= pages.length ? 0 : _index;

    return Scaffold(
      body: IndexedStack(index: safeIndex, children: pages),
      bottomNavigationBar: NavigationBar(
        selectedIndex: safeIndex,
        onDestinationSelected: (i) {
          // Dashboard stats (pending leaves, today's classes, etc.) are
          // mutated from other tabs entirely — Leaves, Fees, Salary — none
          // of which know this shared summary exists. Since IndexedStack
          // keeps the dashboard mounted forever, the only reliable place
          // to force a refetch is right here, on the way back to it.
          if (i == 0 && i != safeIndex) {
            if (isAdmin) {
              ref.invalidate(overallSummaryProvider);
            } else {
              ref.invalidate(todayClassesProvider);
              ref.invalidate(myCoachProfileProvider);
            }
          }
          setState(() => _index = i);
        },
        destinations: destinations,
      ),
    );
  }
}
