import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../features/activities/presentation/activities_screen.dart';
import '../../features/analytics/presentation/analytics_screen.dart';
import '../../features/attendance/presentation/class_attendance_screen.dart';
import '../../features/attendance/presentation/coach_attendance_screen.dart';
import '../../features/attendance/presentation/coach_checkinout_screen.dart';
import '../../features/auth/data/auth_controller.dart';
import '../../features/auth/presentation/login_screen.dart';
import '../../features/classes/presentation/batches_screen.dart';
import '../../features/classes/presentation/substitutions_screen.dart';
import '../../features/dashboard/presentation/home_shell.dart';
import '../../features/fees/presentation/fees_screen.dart';
import '../../features/leaves/presentation/leave_detail_screen.dart';
import '../../features/leaves/presentation/leaves_screen.dart';
import '../../features/locations/presentation/locations_screen.dart';
import '../../features/notifications/presentation/notifications_screen.dart';
import '../../features/salary/presentation/salary_screen.dart';
import '../../features/students/presentation/student_detail_screen.dart';
import '../../features/users/presentation/users_screen.dart';

/// Bridges Riverpod's async auth state into a Listenable so GoRouter's
/// `refreshListenable` re-evaluates `redirect` whenever sign-in state
/// changes, without the router itself needing to be a Riverpod widget.
class _AuthRefreshNotifier extends ChangeNotifier {
  _AuthRefreshNotifier(Ref ref) {
    ref.listen(authControllerProvider, (_, __) => notifyListeners());
  }
}

final routerProvider = Provider<GoRouter>((ref) {
  final authNotifier = _AuthRefreshNotifier(ref);

  return GoRouter(
    initialLocation: '/splash',
    refreshListenable: authNotifier,
    redirect: (context, state) {
      final authState = ref.read(authControllerProvider);
      final atSplash = state.matchedLocation == '/splash';
      final atLogin = state.matchedLocation == '/login';

      if (authState.isLoading) {
        return atSplash ? null : '/splash';
      }
      final session = authState.valueOrNull;
      if (session == null) {
        return atLogin ? null : '/login';
      }
      if (atLogin || atSplash) return '/home';
      return null;
    },
    routes: [
      GoRoute(path: '/splash', builder: (context, state) => const _SplashScreen()),
      GoRoute(path: '/login', builder: (context, state) => const LoginScreen()),
      GoRoute(path: '/home', builder: (context, state) => const HomeShell()),
      GoRoute(
        path: '/students/:id',
        builder: (context, state) => StudentDetailScreen(studentId: state.pathParameters['id']!),
      ),
      GoRoute(
        path: '/classes/:id/attendance',
        builder: (context, state) => ClassAttendanceScreen(classId: state.pathParameters['id']!),
      ),
      GoRoute(
        path: '/coach-checkin/:classId',
        builder: (context, state) => CoachCheckInOutScreen(classId: state.pathParameters['classId']!),
      ),
      GoRoute(path: '/leaves', builder: (context, state) => const LeavesScreen()),
      GoRoute(
        path: '/leaves/:id',
        builder: (context, state) => LeaveDetailScreen(leaveId: state.pathParameters['id']!),
      ),
      GoRoute(path: '/fees', builder: (context, state) => const FeesScreen()),
      GoRoute(path: '/salary', builder: (context, state) => const SalaryScreen()),
      GoRoute(path: '/notifications', builder: (context, state) => const NotificationsScreen()),
      GoRoute(path: '/analytics', builder: (context, state) => const AnalyticsScreen()),
      GoRoute(path: '/substitutions', builder: (context, state) => const SubstitutionsScreen()),
      GoRoute(path: '/coach-attendance', builder: (context, state) => const CoachAttendanceScreen()),
      GoRoute(path: '/activities', builder: (context, state) => const ActivitiesScreen()),
      GoRoute(path: '/batches', builder: (context, state) => const BatchesScreen()),
      GoRoute(path: '/locations', builder: (context, state) => const LocationsScreen()),
      GoRoute(path: '/users', builder: (context, state) => const UsersScreen()),
    ],
  );
});

class _SplashScreen extends StatelessWidget {
  const _SplashScreen();

  @override
  Widget build(BuildContext context) {
    return const Scaffold(body: Center(child: CircularProgressIndicator()));
  }
}
